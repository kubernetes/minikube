/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

const longDescription = `Outputs minikube shell completion for the given shell (bash, zsh or fish)

	This depends on the bash-completion binary.  Example installation instructions:
	OS X:
		$ brew install bash-completion
		$ source $(brew --prefix)/etc/bash_completion
		$ minikube completion bash > $(brew --prefix)/etc/bash_completion.d/minikube  # for bash users
		$ minikube completion zsh | tee "${fpath[1]}/_minikube"  # for zsh users
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish  # for fish users
	Ubuntu:
		$ apt-get install bash-completion
		$ dir="${BASH_COMPLETION_USER_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion}/completions"
		$ mkdir -p "${dir}" && minikube completion bash > "${dir}/minikube"  # for bash users
		$ minikube completion zsh | tee "${fpath[1]}/_minikube"  # for zsh users
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish  # for fish users

	Then restart the shell

	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2
	Note for fish users: [2] please refer to this docs for more details https://fishshell.com/docs/current/#tab-completion
`

const boilerPlate = `
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var completionCmd = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Generate command completion for a shell",
	Long:  longDescription,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "Usage: minikube completion SHELL")
		}
		if args[0] == "bash" {
			err := GenerateBashCompletion(os.Stdout, cmd.Parent())
			if err != nil {
				exit.Error(reason.InternalCompletion, "bash completion failed", err)
			}
		} else if args[0] == "zsh" {
			err := GenerateZshCompletion(os.Stdout, cmd.Parent())
			if err != nil {
				exit.Error(reason.InternalCompletion, "zsh completion failed", err)
			}
		} else if args[0] == "fish" {
			err := GenerateFishCompletion(os.Stdout, cmd.Parent())
			if err != nil {
				exit.Error(reason.InternalCompletion, "fish completion failed", err)
			}
		} else {
			exit.Message(reason.Usage, "Sorry, completion support is not yet implemented for {{.name}}", out.V{"name": args[0]})
		}
	},
}

// GenerateBashCompletion generates the completion for the bash shell
func GenerateBashCompletion(w io.Writer, cmd *cobra.Command) error {
	_, err := w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	err = cmd.GenBashCompletion(w)
	if err != nil {
		return errors.Wrap(err, "Error generating bash completion")
	}

	return nil
}

// GenerateZshCompletion generates the completion for the zsh shell
func GenerateZshCompletion(w io.Writer, cmd *cobra.Command) error {
	zshAutoloadTag := `#compdef _minikube minikube
`

	_, err := w.Write([]byte(zshAutoloadTag))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	err = cmd.GenZshCompletion(w)
	if err != nil {
		return err
	}

	return nil
}

// GenerateFishCompletion generates the completion for the bash shell
func GenerateFishCompletion(w io.Writer, cmd *cobra.Command) error {
	_, err := w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	err = cmd.GenFishCompletion(w, true)
	if err != nil {
		return errors.Wrap(err, "Error generating fish completion")
	}

	return nil
}
