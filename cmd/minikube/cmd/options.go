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

package cmd

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/minikube/pkg/minikube/out"
)

// optionsCmd represents the options command
var optionsCmd = &cobra.Command{
	Use:    "options",
	Short:  "Show a list of global command-line options (applies to all commands).",
	Long:   "Show a list of global command-line options (applies to all commands).",
	Hidden: true,
	Run:    runOptions,
}

// runOptions handles the executes the flow of "minikube options"
func runOptions(cmd *cobra.Command, args []string) {
	out.String("The following options can be passed to any command:\n\n", false)
	cmd.Root().PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		out.String(flagUsage(flag), false)
	})
}

func flagUsage(flag *pflag.Flag) string {
	x := new(bytes.Buffer)

	if flag.Hidden {
		return ""
	}

	format := "--%s=%s: %s\n"

	if flag.Value.Type() == "string" {
		format = "--%s='%s': %s\n"
	}

	if len(flag.Shorthand) > 0 {
		format = "  -%s, " + format
	} else {
		format = "   %s   " + format
	}

	fmt.Fprintf(x, format, flag.Shorthand, flag.Name, flag.DefValue, flag.Usage)

	return x.String()
}
