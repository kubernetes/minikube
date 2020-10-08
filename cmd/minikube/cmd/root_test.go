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
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/tests"
)

func runCommand(f func(*cobra.Command, []string)) {
	cmd := cobra.Command{}
	var args []string
	f(&cmd, args)
}

func TestPreRunDirectories(t *testing.T) {
	// Make sure we create the required directories.
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	runCommand(RootCmd.PersistentPreRun)

	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Fatalf("Directory %s does not exist.", dir)
		}
	}
}

func initTestConfig(config string) error {
	viper.SetConfigType("json")
	r := bytes.NewReader([]byte(config))
	return viper.ReadConfig(r)
}

func TestViperConfig(t *testing.T) {
	defer viper.Reset()
	err := initTestConfig(`{ "v": "999" }`)
	if viper.GetString("v") != "999" || err != nil {
		t.Fatalf("Viper did not read test config file: %v", err)
	}
}
