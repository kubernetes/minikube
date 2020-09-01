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
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/shell"
)

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
		unsetSuffix:    "\n",
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

var (
	defaultSh              = "bash"
	defaultShell shellData = shellConfigMap[defaultSh]
)

// ForceShell forces a shell name
var ForceShell string

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

// UnsetScript writes out a shell-compatible unset script
func UnsetScript(ec EnvConfig, w io.Writer, vars []string) error {
	var sb strings.Builder
	shellCfg := ec.getShell()
	pfx, sfx, delim := shellCfg.unsetPrefix, shellCfg.unsetSuffix, shellCfg.unsetDelimiter
	switch ec.Shell {
	case "cmd", "emacs", "fish":
		break
	case "powershell":
		vars = []string{strings.Join(vars, " Env:\\\\")}
	default:
		vars = []string{strings.Join(vars, " ")}
	}
	for _, v := range vars {
		if _, err := sb.WriteString(fmt.Sprintf("%s%s%s%s", pfx, v, delim, sfx)); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}
