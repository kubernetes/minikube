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

var shellConfigMap = map[string]map[string]string{
	"fish": {
		"Prefix":      "set -gx ",
		"Suffix":      "\";\n", // semi-colon required for fish 2.7
		"Delimiter":   " \"",
		"UnsetPrefix": "set -e ",
		"UnsetSuffix": ";\n",
	},
	"powershell": {
		"Prefix":      "$Env:",
		"Suffix":      "\"\n",
		"Delimiter":   " = \"",
		"UnsetPrefix": `Remove-Item Env:\\`,
		"UnsetSuffix": "\n",
	},
	"cmd": {
		"Prefix":      "SET ",
		"Suffix":      "\n",
		"Delimiter":   "=",
		"UnsetPrefix": "SET ",
		"UnsetSuffix": "\n",
		"setDelim":    "=",
	},
	"emacs": {
		"Prefix":      "(setenv \"",
		"Suffix":      "\")\n",
		"Delimiter":   "\" \"",
		"UnsetPrefix": "(setenv \"",
		"UnsetSuffix": ")\n",
		"UnsetDelim":  "\" nil",
	},
	"bash": {
		"Prefix":      "export ",
		"Suffix":      "\"\n",
		"Delimiter":   "=\"",
		"UnsetPrefix": "unset ",
		"UnsetSuffix": "\n",
	},
	"none": {
		"Prefix":    "",
		"Suffix":    "\n",
		"Delimiter": "=",
	},
}

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
	var usageHintMap = map[string]string{
		"bash": fmt.Sprintf(`
# %s
# eval $(%s)
`, usgPlz, usgCmd),
		"fish": fmt.Sprintf(`
# %s
# %s | source
`, usgPlz, usgCmd),
		"powershell": fmt.Sprintf(`# %s
# & %s | Invoke-Expression
`, usgPlz, usgCmd),
		"cmd": fmt.Sprintf(`REM %s
REM @FOR /f "tokens=*" %%i IN ('%s') DO @%%i
`, usgPlz, usgCmd),
		"emacs": fmt.Sprintf(`;; %s
;; (with-temp-buffer (shell-command "%s" (current-buffer)) (eval-buffer))
`, usgPlz, usgCmd),
	}

	hint, ok := usageHintMap[sh]
	if !ok {
		return usageHintMap["bash"]
	}
	return hint
}

// CfgSet generates context variables for shell
func CfgSet(ec EnvConfig, plz, cmd string) *Config {

	shellKey, s := ec.Shell, &Config{}
	if _, ok := shellConfigMap[shellKey]; !ok {
		shellKey = "bash"
	}
	shellParams := shellConfigMap[shellKey]
	s.Suffix, s.Prefix, s.Delimiter = shellParams["Suffix"], shellParams["Prefix"], shellParams["Delimiter"]

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
	shCfg := shellConfigMap[ec.Shell]
	pfx, sfx, delim := shCfg["Prefix"], shCfg["Suffix"], shCfg["Delimiter"]
	switch ec.Shell {
	case "cmd", "emacs", "fish":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s%s", pfx, v, delim, sfx))
		}
	case "powershell":
		sb.WriteString(fmt.Sprintf("%s%s%s", pfx, strings.Join(vars, " Env:\\\\"), sfx))
	default:
		sb.WriteString(fmt.Sprintf("%s%s%s", pfx, strings.Join(vars, " "), sfx))
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}
