/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package generate

import (
	"strings"

	"github.com/spf13/cobra"
)

type rewrite struct {
	flag        string
	description string
}

// rewriteFlags rewrites flags that are dependent on operating system
// for example, for `minikube start`, the description of --driver
// outputs possible drivers for the operating system
func rewriteFlags(command *cobra.Command, contents string) string {
	rewrites := map[string][]rewrite{
		"start": []rewrite{{
			flag:        "--driver",
			description: "Used to specify the driver to run kubernetes in. The list of available drivers depends on operating system.",
		}},
	}
	rws, ok := rewrites[command.Name()]
	if !ok {
		return contents
	}
	for _, r := range rws {
		contents = rewriteFlag(contents, r.flag, r.description)
	}
	return contents
}

func rewriteFlag(contents, flag, description string) string {
	lines := strings.Split(contents, "\n")
	for i, l := range lines {
		if strings.Contains(l, flag) {
			// docs start with a prefix of 6 spaces
			replacement := "      "
			replacement += flag
			// there are 36 spaces between the start of the flag name
			// and the description
			spacesBetween := 36 - len(flag)
			for i := 0; i < spacesBetween; i++ {
				replacement += " "
			}
			replacement += description
			lines[i] = replacement
		}
	}
	return strings.Join(lines, "\n")
}
