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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rewriteLogFile resets the magic "log_file" flag
// otherwise it starts out under $TMP somewhere
// and shows up as Changed, to default value ""
func rewriteLogFile() error {
	flag := pflag.Lookup("log_file")
	if flag != nil {
		// don't show the default value
		err := pflag.Set("log_file", "")
		if err != nil {
			return err
		}
	}
	return nil
}

type rewrite struct {
	flag       string
	usage      string
	defaultVal string
}

// rewriteFlags rewrites flags that are dependent on operating system
// for example, for `minikube start`, the usage of --driver
// outputs possible drivers for the operating system
func rewriteFlags(command *cobra.Command) error {
	rewrites := map[string][]rewrite{
		"start": {{
			flag:  "driver",
			usage: "Used to specify the driver to run Kubernetes in. The list of available drivers depends on operating system.",
		}, {
			flag:  "mount-string",
			usage: "The argument to pass the minikube mount command on start.",
		}},
	}
	rws, ok := rewrites[command.Name()]
	if !ok {
		return nil
	}
	for _, r := range rws {
		flag := command.Flag(r.flag)
		if flag == nil {
			return fmt.Errorf("--%s is not a valid flag for %s", r.flag, command.Name())
		}
		flag.Usage = r.usage
		flag.DefValue = r.defaultVal
	}
	return nil
}
