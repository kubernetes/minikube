/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/perf"
)

var rootCmd = &cobra.Command{
	Use:           "mkcmp [path to first binary] [path to second binary]",
	Short:         "mkcmp is used to compare performance of two minikube binaries",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validateArgs(args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		binaries, err := retrieveBinaries(args)
		if err != nil {
			return err
		}
		return perf.CompareMinikubeStart(context.Background(), os.Stdout, binaries)
	},
}

func init() {
	// flag.Parse()
}

func validateArgs(args []string) error {
	if len(args) != 2 {
		return errors.New("mkcmp requires two minikube binaries to compare: mkcmp [path to first binary] [path to second binary]")
	}
	return nil
}

func retrieveBinaries(args []string) ([]*perf.Binary, error) {
	binaries := []*perf.Binary{}
	for _, a := range args {
		binary, err := perf.NewBinary(a)
		if err != nil {
			return nil, err
		}
		binaries = append(binaries, binary)
	}
	return binaries, nil
}

// Execute runs the mkcmp command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
