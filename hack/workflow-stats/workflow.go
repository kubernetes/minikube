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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// StepTimeouts maps a workflow step match key to timeout minutes.
type StepTimeouts map[string]int

// ParseWorkflowYAML parses a GitHub Actions workflow YAML document.
func ParseWorkflowYAML(data []byte) (*yaml.Node, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	return &root, nil
}

// UpdateTimeouts updates timeout-minutes on matching workflow steps.
func UpdateTimeouts(data []byte, timeouts StepTimeouts) ([]byte, error) {
	root, err := ParseWorkflowYAML(data)
	if err != nil {
		return nil, err
	}
	if err := updateWorkflowTimeouts(root, timeouts); err != nil {
		return nil, err
	}
	return EncodeYAML(root)
}

func updateWorkflowTimeouts(root *yaml.Node, timeouts StepTimeouts) error {
	doc, err := workflowDocument(root)
	if err != nil {
		return err
	}

	jobsNode, err := Get(doc, "jobs")
	if err != nil {
		return err
	}
	jobs, err := Entries(jobsNode)
	if err != nil {
		return fmt.Errorf("jobs: %w", err)
	}

	for _, job := range jobs {
		stepsNode, err := Get(job.Value, "steps")
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("job %q: %w", job.Key, err)
		}
		steps, err := Elements(stepsNode)
		if err != nil {
			return fmt.Errorf("job %q steps: %w", job.Key, err)
		}

		for _, stepNode := range steps {
			matchKey, err := stepMatchKey(stepNode)
			if errors.Is(err, ErrNotFound) {
				continue
			}
			if err != nil {
				return fmt.Errorf("job %q step: %w", job.Key, err)
			}

			timeout, ok := timeouts[matchKey]
			if !ok {
				continue
			}
			if err := SetScalar(stepNode, "timeout-minutes", strconv.Itoa(timeout)); err != nil {
				return fmt.Errorf("job %q step %q: %w", job.Key, matchKey, err)
			}
		}
	}
	return nil
}

// Check if the root node is a document node and extract it, else throw error
func workflowDocument(root *yaml.Node) (*yaml.Node, error) {
	if root == nil || len(root.Content) == 0 {
		return nil, fmt.Errorf("empty workflow document")
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("workflow document is %v, want mapping", doc.Kind)
	}
	return doc, nil
}

func stepMatchKey(stepNode *yaml.Node) (string, error) {
	if err := wantKind(stepNode, yaml.MappingNode); err != nil {
		return "", err
	}

	if name, err := optionalString(stepNode, "name"); err != nil {
		return "", err
	} else if name != "" {
		return name, nil
	}
	if uses, err := optionalString(stepNode, "uses"); err != nil {
		return "", err
	} else if uses != "" {
		return "Run " + uses, nil
	}
	if run, err := optionalString(stepNode, "run"); err != nil {
		return "", err
	} else if run != "" {
		firstLine := strings.SplitN(strings.TrimSpace(run), "\n", 2)[0]
		return "Run " + firstLine, nil
	}
	return "", ErrNotFound
}

func optionalString(node *yaml.Node, key string) (string, error) {
	value, err := GetString(node, key)
	if errors.Is(err, ErrNotFound) {
		return "", nil
	}
	return value, err
}
