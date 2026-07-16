/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ErrNotFound means a mapping key was not present.
var ErrNotFound = errors.New("not found")

// MapEntry is a key-value pair from a YAML mapping node.
type MapEntry struct {
	Key   string
	Value *yaml.Node
}

// Entries returns ordered key-value pairs from a YAML mapping node.
func Entries(node *yaml.Node) ([]MapEntry, error) {
	if err := wantKind(node, yaml.MappingNode); err != nil {
		return nil, err
	}

	entries := make([]MapEntry, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		entries = append(entries, MapEntry{
			Key:   node.Content[i].Value,
			Value: node.Content[i+1],
		})
	}
	return entries, nil
}

// Elements returns ordered children from a YAML sequence node.
func Elements(node *yaml.Node) ([]*yaml.Node, error) {
	if err := wantKind(node, yaml.SequenceNode); err != nil {
		return nil, err
	}
	return node.Content, nil
}

// Get returns a mapping value node by key.
func Get(node *yaml.Node, key string) (*yaml.Node, error) {
	if err := wantKind(node, yaml.MappingNode); err != nil {
		return nil, fmt.Errorf("get %q: %w", key, err)
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], nil
		}
	}
	return nil, fmt.Errorf("get %q: %w", key, ErrNotFound)
}

// GetString returns the scalar string value for a mapping key.
func GetString(node *yaml.Node, key string) (string, error) {
	val, err := Get(node, key)
	if err != nil {
		return "", err
	}
	if err := wantKind(val, yaml.ScalarNode); err != nil {
		return "", fmt.Errorf("get %q: %w", key, err)
	}
	return val.Value, nil
}

// SetScalar updates or appends a scalar key-value pair in a mapping node.
func SetScalar(node *yaml.Node, key, value string) error {
	if err := wantKind(node, yaml.MappingNode); err != nil {
		return fmt.Errorf("set %q: %w", key, err)
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			if err := wantKind(node.Content[i+1], yaml.ScalarNode); err != nil {
				return fmt.Errorf("set %q: %w", key, err)
			}
			node.Content[i+1].Kind = yaml.ScalarNode
			node.Content[i+1].Tag = ""
			node.Content[i+1].Value = value
			return nil
		}
	}

	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
	return nil
}

// EncodeYAML encodes a YAML document with the repository's standard indent.
func EncodeYAML(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		_ = enc.Close()
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func wantKind(node *yaml.Node, kind yaml.Kind) error {
	if node == nil {
		return fmt.Errorf("got nil node, want %s", yamlKind(kind))
	}
	if node.Kind != kind {
		return fmt.Errorf("got %s, want %s", yamlKind(node.Kind), yamlKind(kind))
	}
	return nil
}

func yamlKind(kind yaml.Kind) string {
	switch kind {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return fmt.Sprintf("yaml kind %d", kind)
	}
}
