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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

var yamlExampleConfig = []byte(`v: 999
alsologtostderr: true
log_dir: "/etc/hosts"
`)

type configTest struct {
	Name          string
	EnvValue      string
	ConfigValue   string
	FlagValue     string
	ExpectedValue string
}

var configTests = []configTest{
	{
		Name:          "v",
		ExpectedValue: "0",
	},
	{
		Name:          "v",
		ConfigValue:   "999",
		ExpectedValue: "999",
	},
	{
		Name:          "v",
		FlagValue:     "0",
		ExpectedValue: "0",
	},
	{
		Name:          "v",
		EnvValue:      "123",
		ExpectedValue: "123",
	},
	{
		Name:          "v",
		FlagValue:     "3",
		ExpectedValue: "3",
	},
	// Flag should override config and env
	{
		Name:          "v",
		FlagValue:     "3",
		ConfigValue:   "222",
		EnvValue:      "888",
		ExpectedValue: "3",
	},
	// Env should override config
	{
		Name:          "v",
		EnvValue:      "2",
		ConfigValue:   "999",
		ExpectedValue: "2",
	},
	// Config should not override flags not on whitelist
	{
		Name:          "log_backtrace_at",
		ConfigValue:   ":3",
		ExpectedValue: ":0",
	},
	// Env should not override flags not on whitelist
	{
		Name:          "log_backtrace_at",
		EnvValue:      ":2",
		ExpectedValue: ":0",
	},
}

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

func initTestConfig(config string) {
	viper.SetConfigType("yml")
	r := bytes.NewReader([]byte(config))
	viper.ReadConfig(r)
}

func TestViperConfig(t *testing.T) {
	defer viper.Reset()
	initTestConfig("v: 999")
	if viper.GetString("v") != "999" {
		t.Fatalf("Viper did not read test config file")
	}
}

func getEnvVarName(name string) string {
	return constants.MinikubeEnvPrefix + "_" + strings.ToUpper(name)
}

func setValues(tt configTest) {
	if tt.FlagValue != "" {
		pflag.Set(tt.Name, tt.FlagValue)
	}
	if tt.EnvValue != "" {
		os.Setenv(getEnvVarName(tt.Name), tt.EnvValue)
	}
	if tt.ConfigValue != "" {
		initTestConfig(tt.Name + ": " + tt.ConfigValue)
	}
}

func unsetValues(tt configTest) {
	var f = pflag.Lookup(tt.Name)
	f.Value.Set(f.DefValue)
	f.Changed = false

	os.Unsetenv(getEnvVarName(tt.Name))

	viper.Reset()
}

func TestViperAndFlags(t *testing.T) {
	for _, tt := range configTests {
		fmt.Println(tt)
		setValues(tt)
		setupViper()
		var actual = pflag.Lookup(tt.Name).Value.String()
		if actual != tt.ExpectedValue {
			t.Errorf("pflag.Value(%s) => %s, wanted %s [%+v]", tt.Name, actual, tt.ExpectedValue, tt)
		}
		unsetValues(tt)
	}
}
