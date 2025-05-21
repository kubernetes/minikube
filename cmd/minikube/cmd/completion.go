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
	"bytes"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

const longDescription = `Outputs minikube shell completion for the given shell (bash, zsh, fish or powershell)

	This depends on the bash-completion binary.  Example installation instructions:
	OS X:
		$ brew install bash-completion
		$ source $(brew --prefix)/etc/bash_completion
		$ minikube completion bash > ~/.minikube-completion  # for bash users
		$ minikube completion zsh > ~/.minikube-completion  # for zsh users
		$ source ~/.minikube-completion
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish # for fish users
	Ubuntu:
		$ apt-get install bash-completion
		$ source /etc/bash_completion
		$ source <(minikube completion bash) # for bash users
		$ source <(minikube completion zsh) # for zsh users
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish # for fish users

	Additionally, you may want to output the completion to a file and source in your .bashrc

	Windows:
		## Save completion code to a script and execute in the profile
		PS> minikube completion powershell > $HOME\.minikube-completion.ps1
		PS> Add-Content $PROFILE '. $HOME\.minikube-completion.ps1'

		## Execute completion code in the profile
		PS> Add-Content $PROFILE 'if (Get-Command minikube -ErrorAction SilentlyContinue) {
		        minikube completion powershell | Out-String | Invoke-Expression
		    }'

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
	Run: func(_ *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "Usage: minikube completion SHELL")
		}
		if args[0] != "bash" && args[0] != "zsh" && args[0] != "fish" && args[0] != "powershell" {
			exit.Message(reason.Usage, "Sorry, completion support is not yet implemented for {{.name}}", out.V{"name": args[0]})
		}
	},
}

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "bash completion.",
	Long:  "Generate command completion for bash.",
	Run: func(cmd *cobra.Command, _ []string) {
		err := GenerateBashCompletion(os.Stdout, cmd.Root())
		if err != nil {
			exit.Error(reason.InternalCompletion, "bash completion failed", err)
		}
	},
}

var zshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "zsh completion.",
	Long:  "Generate command completion for zsh.",
	Run: func(cmd *cobra.Command, _ []string) {
		err := GenerateZshCompletion(os.Stdout, cmd.Root())
		if err != nil {
			exit.Error(reason.InternalCompletion, "zsh completion failed", err)
		}
	},
}

var fishCmd = &cobra.Command{
	Use:   "fish",
	Short: "fish completion.",
	Long:  "Generate command completion for fish .",
	Run: func(cmd *cobra.Command, _ []string) {
		err := GenerateFishCompletion(os.Stdout, cmd.Root())
		if err != nil {
			exit.Error(reason.InternalCompletion, "fish completion failed", err)
		}
	},
}

var powershellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "powershell completion.",
	Long:  "Generate command completion for PowerShell.",
	Run: func(cmd *cobra.Command, _ []string) {
		err := GeneratePowerShellCompletion(os.Stdout, cmd.Root())
		if err != nil {
			exit.Error(reason.InternalCompletion, "powershell completion failed", err)
		}
	},
}

func init() {
	completionCmd.AddCommand(bashCmd)
	completionCmd.AddCommand(zshCmd)
	completionCmd.AddCommand(fishCmd)
	completionCmd.AddCommand(powershellCmd)
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
	zshAutoloadTag := `#compdef minikube
`

	zshInitialization := `
__minikube_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__minikube_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__minikube_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__minikube_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__minikube_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__minikube_declare() {
	if [ "$1" == "-F" ]; then
		whence -w "$@"
	else
		builtin declare "$@"
	fi
}
__minikube_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__minikube_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__minikube_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__minikube_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__minikube_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character
		printf %q "${1:1}"
	else
		printf %q "$1"
	fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__minikube_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__minikube_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__minikube_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__minikube_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__minikube_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__minikube_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/__minikube_declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__minikube_type/g" \
	-e "s/aliashash\[\"\([a-z]*\)\"\]/aliashash[\1]/g" \
	<<'BASH_COMPLETION_EOF'
`

	_, err := w.Write([]byte(zshAutoloadTag))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(zshInitialization))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = cmd.GenBashCompletion(buf)
	if err != nil {
		return errors.Wrap(err, "Error generating zsh completion")
	}
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	zshTail := `
BASH_COMPLETION_EOF
}
__minikube_bash_source <(__minikube_convert_bash_to_zsh)
`
	_, err = w.Write([]byte(zshTail))
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

// GeneratePowerShellCompletion generates the completion for the PowerShell
func GeneratePowerShellCompletion(w io.Writer, cmd *cobra.Command) error {
	_, err := w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	err = cmd.GenPowerShellCompletionWithDesc(w)
	if err != nil {
		return errors.Wrap(err, "Error generating powershell completion")
	}

	return nil
}
