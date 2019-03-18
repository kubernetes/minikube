// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

// magicNotFoundMessage is the string that the CLI special cases to mean
// that a given registry domain wasn't found.
const (
	magicNotFoundMessage = "credentials not found in native keychain"
)

// runner allows us to swap out how we "Run" os/exec commands.
type runner interface {
	Run(*exec.Cmd) error
}

// defaultRunner implements runner by just calling Run().
type defaultRunner struct{}

// Run implements runner.
func (dr *defaultRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// helper executes the named credential helper against the given domain.
type helper struct {
	name   string
	domain name.Registry

	// We add this layer of indirection to facilitate unit testing.
	r runner
}

// helperOutput is the expected JSON output form of a credential helper
// (or at least these are the fields that we care about).
type helperOutput struct {
	Username string
	Secret   string
}

// Authorization implements Authenticator.
func (h *helper) Authorization() (string, error) {
	helperName := fmt.Sprintf("docker-credential-%s", h.name)
	// We want to execute:
	//   echo -n {domain} | docker-credential-{name} get
	cmd := exec.Command(helperName, "get")

	// Some keychains expect a scheme:
	// https://github.com/bazelbuild/rules_docker/issues/111
	cmd.Stdin = strings.NewReader(fmt.Sprintf("https://%v", h.domain))

	var out bytes.Buffer
	cmd.Stdout = &out
	cmdErr := h.r.Run(cmd)

	// If we see this specific message, it means the domain wasn't found
	// and we should fall back on anonymous auth.
	output := strings.TrimSpace(out.String())
	if output == magicNotFoundMessage {
		return Anonymous.Authorization()
	}

	// Any other output should be parsed as JSON and the Username / Secret
	// fields used for Basic authentication.
	ho := helperOutput{}
	if err := json.Unmarshal([]byte(output), &ho); err != nil {
		if cmdErr != nil {
			// If we failed to parse output, it won't contain Secret, so returning it
			// in an error should be fine.
			return "", fmt.Errorf("invoking %s: %v; output: %s", helperName, cmdErr, output)
		}
		return "", err
	}

	if cmdErr != nil {
		return "", fmt.Errorf("invoking %s: %v", helperName, cmdErr)
	}

	b := Basic{Username: ho.Username, Password: ho.Secret}
	return b.Authorization()
}
