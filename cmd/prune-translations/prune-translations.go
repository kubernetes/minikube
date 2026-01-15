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

// This tool removes stale translations (keys not in strings.txt) from language JSON files.
// Usage: go run cmd/prune-translations/prune-translations.go (from minikube root)

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PruneResult holds the result of pruning a single translation file
type PruneResult struct {
	File     string
	Original int
	Removed  int
	Kept     int
}

// PruneTranslations removes stale translations from JSON files in the given directory.
// It returns results for each processed file and any error encountered.
func PruneTranslations(translationsDir string) ([]PruneResult, error) {
	// Read valid keys from strings.txt
	stringsPath := filepath.Join(translationsDir, "strings.txt")
	data, err := os.ReadFile(stringsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read strings.txt: %w", err)
	}

	var validKeys map[string]interface{}
	if err := json.Unmarshal(data, &validKeys); err != nil {
		return nil, fmt.Errorf("failed to parse strings.txt: %w", err)
	}

	// Process each JSON file
	pattern := filepath.Join(translationsDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob JSON files: %w", err)
	}

	var results []PruneResult
	for _, path := range files {
		result, err := pruneFile(path, validKeys)
		if err != nil {
			return nil, fmt.Errorf("failed to prune %s: %w", path, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// pruneFile removes stale keys from a single translation file
func pruneFile(path string, validKeys map[string]interface{}) (PruneResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PruneResult{}, err
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return PruneResult{}, err
	}

	original := len(translations)
	for key := range translations {
		if _, ok := validKeys[key]; !ok {
			delete(translations, key)
		}
	}

	removed := original - len(translations)
	if removed > 0 {
		output, err := json.MarshalIndent(translations, "", "\t")
		if err != nil {
			return PruneResult{}, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return PruneResult{}, err
		}
		if err := os.WriteFile(path, append(output, '\n'), info.Mode()); err != nil {
			return PruneResult{}, err
		}
	}

	return PruneResult{
		File:     filepath.Base(path),
		Original: original,
		Removed:  removed,
		Kept:     len(translations),
	}, nil
}

func main() {
	results, err := PruneTranslations("translations")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processing translations...\n\n")
	for _, r := range results {
		fmt.Printf("  %s: %d removed, %d kept\n", r.File, r.Removed, r.Kept)
	}
	fmt.Println("\nDone!")
}
