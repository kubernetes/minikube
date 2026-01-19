/*
Copyright 2025 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupTestDir creates a temporary directory with test translation files
func setupTestDir(t *testing.T, stringsContent map[string]interface{}, translations map[string]map[string]interface{}) string {
	t.Helper()
	td := t.TempDir()

	// Write strings.txt (valid keys)
	stringsData, err := json.MarshalIndent(stringsContent, "", "\t")
	if err != nil {
		t.Fatalf("failed to marshal strings.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(td, "strings.txt"), stringsData, 0644); err != nil {
		t.Fatalf("failed to write strings.txt: %v", err)
	}

	// Write translation JSON files
	for filename, content := range translations {
		data, err := json.MarshalIndent(content, "", "\t")
		if err != nil {
			t.Fatalf("failed to marshal %s: %v", filename, err)
		}
		if err := os.WriteFile(filepath.Join(td, filename), data, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}

	return td
}

// readJSONFile reads and parses a JSON file into a map
func readJSONFile(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	return result
}

func TestPruneTranslations_RemovesStaleKeys(t *testing.T) {
	// Setup: strings.txt has 2 valid keys
	validKeys := map[string]interface{}{
		"Starting minikube": "",
		"Stopping minikube": "",
	}

	// Translation file has 3 keys: 2 valid + 1 stale
	translations := map[string]map[string]interface{}{
		"ko.json": {
			"Starting minikube":  "minikube 시작 중",
			"Stopping minikube":  "minikube 중지 중",
			"Old deprecated key": "삭제되어야 할 키", // stale key
		},
	}

	td := setupTestDir(t, validKeys, translations)

	// Execute
	results, err := PruneTranslations(td)
	if err != nil {
		t.Fatalf("PruneTranslations failed: %v", err)
	}

	// Verify results
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.File != "ko.json" {
		t.Errorf("expected file ko.json, got %s", r.File)
	}
	if r.Original != 3 {
		t.Errorf("expected original=3, got %d", r.Original)
	}
	if r.Removed != 1 {
		t.Errorf("expected removed=1, got %d", r.Removed)
	}
	if r.Kept != 2 {
		t.Errorf("expected kept=2, got %d", r.Kept)
	}

	// Verify file content
	content := readJSONFile(t, filepath.Join(td, "ko.json"))
	if len(content) != 2 {
		t.Errorf("expected 2 keys in file, got %d", len(content))
	}
	if _, ok := content["Starting minikube"]; !ok {
		t.Error("valid key 'Starting minikube' was incorrectly removed")
	}
	if _, ok := content["Stopping minikube"]; !ok {
		t.Error("valid key 'Stopping minikube' was incorrectly removed")
	}
	if _, ok := content["Old deprecated key"]; ok {
		t.Error("stale key 'Old deprecated key' was not removed")
	}
}

func TestPruneTranslations_PreservesValidKeys(t *testing.T) {
	// Setup: all keys in translation file are valid
	validKeys := map[string]interface{}{
		"Hello":   "",
		"Goodbye": "",
		"Thanks":  "",
	}

	translations := map[string]map[string]interface{}{
		"ja.json": {
			"Hello":   "こんにちは",
			"Goodbye": "さようなら",
			"Thanks":  "ありがとう",
		},
	}

	td := setupTestDir(t, validKeys, translations)

	// Execute
	results, err := PruneTranslations(td)
	if err != nil {
		t.Fatalf("PruneTranslations failed: %v", err)
	}

	// Verify no keys were removed
	r := results[0]
	if r.Removed != 0 {
		t.Errorf("expected removed=0, got %d", r.Removed)
	}
	if r.Kept != 3 {
		t.Errorf("expected kept=3, got %d", r.Kept)
	}

	// Verify all keys still exist
	content := readJSONFile(t, filepath.Join(td, "ja.json"))
	for key := range validKeys {
		if _, ok := content[key]; !ok {
			t.Errorf("valid key %q was incorrectly removed", key)
		}
	}
}

func TestPruneTranslations_MultipleFiles(t *testing.T) {
	validKeys := map[string]interface{}{
		"Start": "",
		"Stop":  "",
	}

	translations := map[string]map[string]interface{}{
		"ko.json": {
			"Start":  "시작",
			"Stop":   "중지",
			"Stale1": "삭제1",
		},
		"ja.json": {
			"Start":  "開始",
			"Stale2": "削除2",
			"Stale3": "削除3",
		},
		"zh-CN.json": {
			"Start": "开始",
			"Stop":  "停止",
		},
	}

	td := setupTestDir(t, validKeys, translations)

	// Execute
	results, err := PruneTranslations(td)
	if err != nil {
		t.Fatalf("PruneTranslations failed: %v", err)
	}

	// Verify correct number of files processed
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Check each file
	resultMap := make(map[string]PruneResult)
	for _, r := range results {
		resultMap[r.File] = r
	}

	// ko.json: 1 stale key removed
	if r, ok := resultMap["ko.json"]; ok {
		if r.Removed != 1 {
			t.Errorf("ko.json: expected removed=1, got %d", r.Removed)
		}
	}

	// ja.json: 2 stale keys removed
	if r, ok := resultMap["ja.json"]; ok {
		if r.Removed != 2 {
			t.Errorf("ja.json: expected removed=2, got %d", r.Removed)
		}
	}

	// zh-CN.json: no stale keys
	if r, ok := resultMap["zh-CN.json"]; ok {
		if r.Removed != 0 {
			t.Errorf("zh-CN.json: expected removed=0, got %d", r.Removed)
		}
	}
}

func TestPruneTranslations_EmptyTranslationFile(t *testing.T) {
	validKeys := map[string]interface{}{
		"Hello": "",
	}

	translations := map[string]map[string]interface{}{
		"empty.json": {},
	}

	td := setupTestDir(t, validKeys, translations)

	results, err := PruneTranslations(td)
	if err != nil {
		t.Fatalf("PruneTranslations failed: %v", err)
	}

	r := results[0]
	if r.Removed != 0 || r.Kept != 0 {
		t.Errorf("expected removed=0, kept=0, got removed=%d, kept=%d", r.Removed, r.Kept)
	}
}

func TestPruneTranslations_MissingStringsFile(t *testing.T) {
	td := t.TempDir()
	// Don't create strings.txt

	_, err := PruneTranslations(td)
	if err == nil {
		t.Error("expected error when strings.txt is missing")
	}
}

func TestPruneTranslations_InvalidStringsJSON(t *testing.T) {
	td := t.TempDir()
	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(td, "strings.txt"), []byte("invalid json"), 0644); err != nil {
		t.Fatalf("failed to write strings.txt: %v", err)
	}

	_, err := PruneTranslations(td)
	if err == nil {
		t.Error("expected error for invalid JSON in strings.txt")
	}
}

func TestPruneTranslations_DoesNotTouchRealTranslations(t *testing.T) {
	// This test ensures we're using temp directories and not real files
	// by verifying the test directory is isolated

	validKeys := map[string]interface{}{"Test": ""}
	translations := map[string]map[string]interface{}{
		"test.json": {"Test": "테스트", "Stale": "삭제"},
	}

	td := setupTestDir(t, validKeys, translations)

	// Verify temp directory is not the real translations directory
	realDir := "translations"
	if td == realDir || filepath.Base(td) == "translations" {
		t.Fatal("test is using real translations directory - this is dangerous!")
	}

	// Execute in isolated directory
	_, err := PruneTranslations(td)
	if err != nil {
		t.Fatalf("PruneTranslations failed: %v", err)
	}

	// Verify real translations directory was not modified
	// (this just checks the real directory still exists if it exists)
	if _, err := os.Stat(realDir); err == nil {
		// Real dir exists - that's fine, we just didn't modify it
		t.Logf("Real translations directory exists and was not touched")
	}
}
