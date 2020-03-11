/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package cruntime

import (
	"bytes"
	"os/exec"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
)

type mockRunner struct {
	runCmdOutput bytes.Buffer
}

func (m *mockRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	return &command.RunResult{
		Stdout: m.runCmdOutput,
	}, nil
}

// Copy is a convenience method that runs a command to copy a file
func (m *mockRunner) Copy(a assets.CopyableFile) error {
	return nil
}

// Remove is a convenience method that runs a command to remove a file
func (m *mockRunner) Remove(assets.CopyableFile) error {
	return nil
}

func TestMergeRepositoriesJSON(t *testing.T) {
	initial := `{
		"Repositories": {
		  "imageOne": {
			"image:tag": "digest",
			"image@digest": "digest"
		  },
		  "imageTwo": {
			"image2:tag": "digest",
			"image2@digest": "digest"
		  }
		}
	}`

	afterPreload := `{
		"Repositories": {
		  "imageOne": {
			"image:tag": "new_digest",
			"image@digest": "new_digest"
		  },
		  "imageThree": {
			"image2:tag": "digest_three",
			"image2@digest": "digest_three"
		  }
		}
	}`

	expected := `{
		"Repositories": {
		  "imageOne": {
			"image:tag": "new_digest",
			"image@digest": "new_digest"
		  },
		  "imageTwo": {
			"image2:tag": "digest",
			"image2@digest": "digest"
		  },
		  "imageThree": {
			"image2:tag": "digest_three",
			"image2@digest": "digest_three"
		  }
		}
	}`

	runner := &mockRunner{
		runCmdOutput: *bytes.NewBuffer([]byte(afterPreload)),
	}

	driver := &Docker{Runner: runner}
	result, err := driver.mergeRepositoriesJSON(initial)
	if err != nil {
		t.Fatalf("error merging repositories.json file: %v", err)
	}

	if string(result) != expected {
		t.Fatalf("didn't get expected repositories.json. Actual:\n%s\nExpected:\n%s\n", result, expected)
	}

}
