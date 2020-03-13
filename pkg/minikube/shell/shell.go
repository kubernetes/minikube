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

const (
	fishSetPfx   = "set -gx "
	fishSetSfx   = "\";\n" // semi-colon required for fish 2.7
	fishSetDelim = " \""

	fishUnsetPfx = "set -e "
	fishUnsetSfx = ";\n"

	psSetPfx   = "$Env:"
	psSetSfx   = "\"\n"
	psSetDelim = " = \""

	psUnsetPfx = `Remove-Item Env:\\`
	psUnsetSfx = "\n"

	cmdSetPfx   = "SET "
	cmdSetSfx   = "\n"
	cmdSetDelim = "="

	cmdUnsetPfx   = "SET "
	cmdUnsetSfx   = "\n"
	cmdUnsetDelim = "="

	emacsSetPfx   = "(setenv \""
	emacsSetSfx   = "\")\n"
	emacsSetDelim = "\" \""

	emacsUnsetPfx   = "(setenv \""
	emacsUnsetSfx   = ")\n"
	emacsUnsetDelim = "\" nil"

	bashSetPfx   = "export "
	bashSetSfx   = "\"\n"
	bashSetDelim = "=\""

	bashUnsetPfx = "unset "
	bashUnsetSfx = "\n"

	nonePfx   = ""
	noneSfx   = "\n"
	noneDelim = "="
)

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
	s := &Config{
		UsageHint: generateUsageHint(ec.Shell, plz, cmd),
	}

	switch ec.Shell {
	case "fish":
		s.Prefix = fishSetPfx
		s.Suffix = fishSetSfx
		s.Delimiter = fishSetDelim
	case "powershell":
		s.Prefix = psSetPfx
		s.Suffix = psSetSfx
		s.Delimiter = psSetDelim
	case "cmd":
		s.Prefix = cmdSetPfx
		s.Suffix = cmdSetSfx
		s.Delimiter = cmdSetDelim
	case "emacs":
		s.Prefix = emacsSetPfx
		s.Suffix = emacsSetSfx
		s.Delimiter = emacsSetDelim
	case "none":
		s.Prefix = nonePfx
		s.Suffix = noneSfx
		s.Delimiter = noneDelim
		s.UsageHint = ""
	default:
		s.Prefix = bashSetPfx
		s.Suffix = bashSetSfx
		s.Delimiter = bashSetDelim
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
	switch ec.Shell {
	case "fish":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s", fishUnsetPfx, v, fishUnsetSfx))
		}
	case "powershell":
		sb.WriteString(fmt.Sprintf("%s%s%s", psUnsetPfx, strings.Join(vars, " Env:\\\\"), psUnsetSfx))
	case "cmd":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s%s", cmdUnsetPfx, v, cmdUnsetDelim, cmdUnsetSfx))
		}
	case "emacs":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s%s", emacsUnsetPfx, v, emacsUnsetDelim, emacsUnsetSfx))
		}
	case "none":
		sb.WriteString(fmt.Sprintf("%s%s%s", nonePfx, strings.Join(vars, " "), noneSfx))
	default:
		sb.WriteString(fmt.Sprintf("%s%s%s", bashUnsetPfx, strings.Join(vars, " "), bashUnsetSfx))
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}
