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
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

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
		ConfigValue:   `{ "v":"999" }`,
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
		ConfigValue:   `{ "v": "222" }`,
		EnvValue:      "888",
		ExpectedValue: "3",
	},
	// Env should override config
	{
		Name:          "v",
		EnvValue:      "2",
		ConfigValue:   `{ "v": "999" }`,
		ExpectedValue: "2",
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

// Temporarily unsets the env variables for the test cases.
// Returns a function to reset them to their initial values.
func hideEnv(t *testing.T) func(t *testing.T) {
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, constants.MinikubeEnvPrefix) {
			line := strings.Split(env, "=")
			key, val := line[0], line[1]
			envs[key] = val
			t.Logf("TestConfig: Unsetting %s=%s for unit test!", key, val)
			os.Unsetenv(key)
		}
	}
	return func(t *testing.T) {
		for key, val := range envs {
			t.Logf("TestConfig: Finished test, Resetting Env %s=%s", key, val)
			os.Setenv(key, val)
		}
	}
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

func getEnvVarName(name string) string {
	return constants.MinikubeEnvPrefix + "_" + strings.ToUpper(name)
}

func setValues(tt configTest) error {
	if tt.FlagValue != "" {
		if err := pflag.Set(tt.Name, tt.FlagValue); err != nil {
			return errors.Wrap(err, "flag set")
		}
	}
	if tt.EnvValue != "" {
		s := strings.Replace(getEnvVarName(tt.Name), "-", "_", -1)
		os.Setenv(s, tt.EnvValue)
	}
	if tt.ConfigValue != "" {
		if err := initTestConfig(tt.ConfigValue); err != nil {
			return errors.Wrapf(err, "Config %s not read correctly", tt.ConfigValue)
		}
	}
	return nil
}

func unsetValues(name string) error {
	f := pflag.Lookup(name)
	if err := f.Value.Set(f.DefValue); err != nil {
		return errors.Wrapf(err, "set(%s)", f.DefValue)
	}
	f.Changed = false
	os.Unsetenv(getEnvVarName(name))
	viper.Reset()
	return nil
}

func TestViperAndFlags(t *testing.T) {
	restore := hideEnv(t)
	defer restore(t)
	for _, tt := range configTests {
		err := setValues(tt)
		if err != nil {
			t.Fatalf("setValues: %v", err)
		}
		setupViper()
		f := pflag.Lookup(tt.Name)
		if f == nil {
			t.Fatalf("Could not find flag for %s", tt.Name)
		}
		actual := f.Value.String()
		if actual != tt.ExpectedValue {
			t.Errorf("pflag.Value(%s) => %s, wanted %s [%+v]", tt.Name, actual, tt.ExpectedValue, tt)
		}
		// Some flag validation may not accept their default value, such as log_at_backtrace :(
		if err := unsetValues(tt.Name); err != nil {
			t.Logf("unsetValues(%s) failed: %v", tt.Name, err)
		}
	}
}
