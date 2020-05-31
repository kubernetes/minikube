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
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/shell"
)

type shellData struct {
	Prefix         string
	Suffix         string
	Delimiter      string
	UnsetPrefix    string
	UnsetSuffix    string
	UnsetDelimiter string
	usageHint      func(s ...interface{}) string
}

var shellConfigMap = map[string]shellData{
	"fish": shellData{
		Prefix:         "set -gx ",
		Suffix:         "\";\n",
		Delimiter:      " \"",
		UnsetPrefix:    "set -e ",
		UnsetSuffix:    ";\n",
		UnsetDelimiter: "",
		usageHint: func(s ...interface{}) string {
			return fmt.Sprintf(`
			# %s
			# %s | source
			`, s...)
		},
	},
	"powershell": shellData{
		"$Env:",
		"\"\n",
		" = \"",
		`Remove-Item Env:\\`,
		"\n",
		"",
		func(s ...interface{}) string {
			return fmt.Sprintf(`# %s
			# & %s | Invoke-Expression
			`, s...)
		},
	},
	"cmd": shellData{
		"SET ",
		"\n",
		"=",
		"SET ",
		"\n",
		"=",
		func(s ...interface{}) string {
			return fmt.Sprintf(`REM %s
			REM @FOR /f "tokens=*" %%i IN ('%s') DO @%%i
			`, s...)
		},
	},
	"emacs": shellData{
		"(setenv \"",
		"\")\n",
		"\" \"",
		"(setenv \"",
		")\n",
		"\" nil",
		func(s ...interface{}) string {
			return fmt.Sprintf(`;; %s
			;; (with-temp-buffer (shell-command "%s" (current-buffer)) (eval-buffer))
			`, s...)
		},
	},
	"bash": shellData{
		"export ",
		"\"\n",
		"=\"",
		"unset ",
		"\n",
		"",
		func(s ...interface{}) string {
			return fmt.Sprintf(`
			# %s
			# eval $(%s)
			`, s...)
		},
	},
	"none": shellData{
		"",
		"\n",
		"=",
		"",
		"\n",
		"=",
		func(s ...interface{}) string {
			return ""
		},
	},
}

var defaultShell shellData = shellConfigMap["bash"]

// Config represents the shell config
type Config struct {
	Prefix    string
	Delimiter string
	Suffix    string
	UsageHint string
}

var (
	// ForceShell forces a shell name
	ForceShell string
)

// Detect detects user's current shell.
func Detect() (string, error) {
	return shell.Detect()
}

func generateUsageHint(sh, usgPlz, usgCmd string) string {
	shellCfg, ok := shellConfigMap[sh]
	if !ok {
		shellCfg = defaultShell
	}
	return shellCfg.usageHint(usgPlz, usgCmd)
}

// CfgSet generates context variables for shell
func CfgSet(ec EnvConfig, plz, cmd string) *Config {

	shellKey, s := ec.Shell, &Config{}

	shellCfg, ok := shellConfigMap[shellKey]
	if !ok {
		shellCfg = defaultShell
	}
	s.Suffix, s.Prefix, s.Delimiter = shellCfg.Suffix, shellCfg.Prefix, shellCfg.Delimiter

	if shellKey != "none" {
		s.UsageHint = generateUsageHint(ec.Shell, plz, cmd)
	}

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

// UnsetScript writes out a shell-compatible unset script
func UnsetScript(ec EnvConfig, w io.Writer, vars []string) error {
	var sb strings.Builder
	shCfg, ok := shellConfigMap[ec.Shell]
	if !ok {
		shCfg = defaultShell
	}
	pfx, sfx, delim := shCfg.UnsetPrefix, shCfg.UnsetSuffix, shCfg.UnsetDelimiter
	switch ec.Shell {
	case "cmd", "emacs", "fish":
		for _, v := range vars {
			if _, err := sb.WriteString(fmt.Sprintf("%s%s%s%s", pfx, v, delim, sfx)); err != nil {
				return err
			}
		}
	case "powershell":
		if _, err := sb.WriteString(fmt.Sprintf("%s%s%s", pfx, strings.Join(vars, " Env:\\\\"), sfx)); err != nil {
			return err
		}
	default:
		if _, err := sb.WriteString(fmt.Sprintf("%s%s%s", pfx, strings.Join(vars, " "), sfx)); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}
