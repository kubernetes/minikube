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

// Part of this code is heavily inspired/copied by the following file:
// github.com/docker/machine/commands/env.go

package shell

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"text/template"

	"github.com/docker/machine/libmachine/shell"

	"k8s.io/minikube/pkg/minikube/constants"
)

var unsetEnvTmpl = "{{ $root := .}}" +
	"{{ range .Unset }}" +
	"{{ $root.UnsetPrefix }}{{ . }}{{ $root.UnsetDelimiter }}{{ $root.UnsetSuffix }}" +
	"{{ end }}" +
	"{{ range .Set }}" +
	"{{ $root.SetPrefix }}{{ .Env }}{{ $root.SetDelimiter }}{{ .Value }}{{ $root.SetSuffix }}" +
	"{{ end }}"

// Config represents the shell config
type Config struct {
	Prefix    string
	Delimiter string
	Suffix    string
	UsageHint string
}

type shellData struct {
	prefix         string
	suffix         string
	delimiter      string
	unsetPrefix    string
	unsetSuffix    string
	unsetDelimiter string
	usageHint      func(s ...interface{}) string
}

var shellConfigMap = map[string]shellData{
	"fish": {
		prefix:         "set -gx ",
		suffix:         "\";\n",
		delimiter:      " \"",
		unsetPrefix:    "set -e ",
		unsetSuffix:    ";\n",
		unsetDelimiter: "",
		usageHint: func(s ...interface{}) string {
			return fmt.Sprintf(`
# %s
# %s | source
`, s...)
		},
	},
	"powershell": {
		prefix:         "$Env:",
		suffix:         "\"\n",
		delimiter:      " = \"",
		unsetPrefix:    `Remove-Item Env:\\`,
		unsetSuffix:    "\n",
		unsetDelimiter: "",
		usageHint: func(s ...interface{}) string {
			return fmt.Sprintf(`# %s
# & %s | Invoke-Expression
`, s...)
		},
	},
	"cmd": {
		prefix:         "SET ",
		suffix:         "\n",
		delimiter:      "=",
		unsetPrefix:    "SET ",
		unsetSuffix:    "\n",
		unsetDelimiter: "=",
		usageHint: func(s ...interface{}) string {
			return fmt.Sprintf(`REM %s
REM @FOR /f "tokens=*" %%i IN ('%s') DO @%%i
`, s...)
		},
	},
	"emacs": {
		prefix:         "(setenv \"",
		suffix:         "\")\n",
		delimiter:      "\" \"",
		unsetPrefix:    "(setenv \"",
		unsetSuffix:    ")\n",
		unsetDelimiter: "\" nil",
		usageHint: func(s ...interface{}) string {
			return fmt.Sprintf(`;; %s
;; (with-temp-buffer (shell-command "%s" (current-buffer)) (eval-buffer))
`, s...)
		},
	},
	"bash": {
		prefix:         "export ",
		suffix:         "\"\n",
		delimiter:      "=\"",
		unsetPrefix:    "unset ",
		unsetSuffix:    ";\n",
		unsetDelimiter: "",
		usageHint: func(s ...interface{}) string {
			return fmt.Sprintf(`
# %s
# eval $(%s)
`, s...)
		},
	},
	"none": {
		prefix:         "",
		suffix:         "\n",
		delimiter:      "=",
		unsetPrefix:    "",
		unsetSuffix:    "\n",
		unsetDelimiter: "",
		usageHint: func(s ...interface{}) string {
			return ""
		},
	},
}

var defaultSh = "bash"
var defaultShell shellData = shellConfigMap[defaultSh]

var (
	// ForceShell forces a shell name
	ForceShell string
)

// Detect detects user's current shell.
func Detect() (string, error) {
	sh := os.Getenv("SHELL")
	// Don't error out when $SHELL has not been set
	if sh == "" && runtime.GOOS != "windows" {
		return defaultSh, nil
	}
	return shell.Detect()
}

func (c EnvConfig) getShell() shellData {
	shell, ok := shellConfigMap[c.Shell]
	if !ok {
		shell = defaultShell
	}
	return shell
}

func generateUsageHint(ec EnvConfig, usgPlz, usgCmd string) string {
	shellCfg := ec.getShell()
	return shellCfg.usageHint(usgPlz, usgCmd)
}

// CfgSet generates context variables for shell
func CfgSet(ec EnvConfig, plz, cmd string) *Config {
	shellCfg := ec.getShell()
	s := &Config{}
	s.Suffix, s.Prefix, s.Delimiter = shellCfg.suffix, shellCfg.prefix, shellCfg.delimiter
	s.UsageHint = generateUsageHint(ec, plz, cmd)

	return s
}

// EnvConfig encapsulates all external inputs into shell generation
type EnvConfig struct {
	Shell string
}

// SetScript writes out a shell-compatible set script
func SetScript(ec EnvConfig, w io.Writer, envTmpl string, data interface{}) error {
	tmpl := template.Must(template.New("envConfig").Parse(envTmpl))
	return tmpl.Execute(w, data)
}

type unsetConfigItem struct {
	Env, Value string
}
type unsetConfig struct {
	Set            []unsetConfigItem
	Unset          []string
	SetPrefix      string
	SetDelimiter   string
	SetSuffix      string
	UnsetPrefix    string
	UnsetDelimiter string
	UnsetSuffix    string
}

// UnsetScript writes out a shell-compatible unset script
func UnsetScript(ec EnvConfig, w io.Writer, vars []string) error {
	shellCfg := ec.getShell()
	cfg := unsetConfig{
		SetPrefix:      shellCfg.prefix,
		SetDelimiter:   shellCfg.delimiter,
		SetSuffix:      shellCfg.suffix,
		UnsetPrefix:    shellCfg.unsetPrefix,
		UnsetDelimiter: shellCfg.unsetDelimiter,
		UnsetSuffix:    shellCfg.unsetSuffix,
	}
	var tempUnset []string
	for _, env := range vars {
		exEnv := constants.MinikubeExistingPrefix + env
		if v := os.Getenv(exEnv); v == "" {
			cfg.Unset = append(cfg.Unset, env)
		} else {
			cfg.Set = append(cfg.Set, unsetConfigItem{
				Env:   env,
				Value: v,
			})
			tempUnset = append(tempUnset, exEnv)
		}
	}
	cfg.Unset = append(cfg.Unset, tempUnset...)

	tmpl := template.Must(template.New("unsetEnv").Parse(unsetEnvTmpl))
	return tmpl.Execute(w, &cfg)
}
