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

package localkube

import (
	"errors"
	"fmt"
	"io"
)

// LocalKube provides a fully functional Kubernetes cluster running entirely through goroutines
type LocalKube struct {
	Servers
}

func (lk *LocalKube) Add(server Server) {
	lk.Servers = append(lk.Servers, server)
}

func (lk *LocalKube) Run(args []string, out io.Writer) error {
	if len(args) < 2 {
		return errors.New("you must choose start <name>, stop <name>, or status")
	}

	switch args[1] {
	case "start":
		// check if just start
		if len(args) == 2 {
			fmt.Fprintln(out, "Starting LocalKube...")
			lk.StartAll()
			return nil
		} else if len(args) == 3 {
			serverName := args[2]
			fmt.Fprintf(out, "Starting `%s`...\n", serverName)
			return lk.Start(serverName)

		} else {
			return errors.New("start: too many arguments")
		}
	case "stop":
		// check if just stop
		if len(args) == 2 {
			fmt.Fprintln(out, "Stopping LocalKube...")
			lk.StopAll()
			return nil
		} else if len(args) == 3 {
			serverName := args[2]
			fmt.Fprintf(out, "Stopping `%s`...\n", serverName)
			return lk.Stop(serverName)
		}
	case "status":
		fmt.Fprintln(out, "LocalKube Status")
		fmt.Fprintln(out, "################\n")

		fmt.Fprintln(out, "Order\tStatus\tName")
		for num, server := range lk.Servers {
			fmt.Fprintf(out, "%d\t%s\t%s\n", num, server.Status(), server.Name())
		}

		fmt.Fprintln(out)
	}
	return errors.New("you must choose start <name>, stop <name>, or status")
}
