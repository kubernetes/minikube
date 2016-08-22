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
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/constants"
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
	defer os.RemoveAll(tempDir)

	runCommand(RootCmd.PersistentPreRun)

	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Fatalf("Directory %s does not exist.", dir)
		}
	}
}

func getEnvVarName(name string) string {
	return constants.MinikubeEnvPrefix + name
}

func TestEnvVariable(t *testing.T) {
	defer os.Unsetenv("WANTUPDATENOTIFICATION")
	initConfig()
	os.Setenv(getEnvVarName("WANTUPDATENOTIFICATION"), "true")
	if !viper.GetBool("WantUpdateNotification") {
		t.Fatalf("Viper did not respect environment variable")
	}
}

func cleanup() {
	pflag.Set("v", "0")
	pflag.Lookup("v").Changed = false
}

func TestFlagShouldOverrideConfig(t *testing.T) {
	defer cleanup()
	viper.Set("v", "1337")
	pflag.Set("v", "100")
	setFlagsUsingViper()
	if viper.GetInt("v") != 100 {
		viper.Debug()
		t.Fatal("Value from viper config overrode explicit flag value")
	}
}

func TestConfigShouldOverrideDefault(t *testing.T) {
	defer cleanup()
	viper.Set("v", "1337")
	setFlagsUsingViper()
	if viper.GetInt("v") != 1337 {
		viper.Debug()
		t.Fatalf("Value from viper config did not override default flag value")
	}
}

func TestFallbackToDefaultFlag(t *testing.T) {
	setFlagsUsingViper()

	if viper.GetInt("stderrthreshold") != 2 {
		t.Logf("stderrthreshold %s", viper.GetInt("stderrthreshold"))
		t.Fatalf("The default flag value was overwritten")
	}

	if viper.GetString("log-flush-frequency") != "5s" {
		t.Logf("log flush frequency: %s", viper.GetString("log-flush-frequency"))
		t.Fatalf("The default flag value was overwritten")
	}
}
