/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package cruntimeinstaller

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

// for escaping systemd template specifiers (e.g. '%i'), which are not supported by minikube
var systemdSpecifierEscaper = strings.NewReplacer("%", "%%")

func rootFileSystemType(rnr runner.Runner) (string, error) {
	fs, err := rnr.RunCmd(exec.Command("bash", "-c", "df --output=fstype / | tail -n 1"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fs.Stdout.String()), nil
}

// updateUnit efficiently updates a systemd unit file
func updateUnit(rnr runner.Runner, name string, content string, dst string) error {
	log.Infof("Updating %s unit: %s ...", name, dst)

	// TODO: find a better place for this
	// Make sure we have a docker.service in the first place..
	// otherwise subsequent logic will fail
	if _, err := rnr.RunCmd(exec.Command("bash", "-c", fmt.Sprintf("sudo mkdir -p %s && sudo touch %s && sudo systemctl -f daemon-reload", path.Dir(dst), dst))); err != nil {
		return err
	}

	if _, err := rnr.RunCmd(exec.Command("bash", "-c", fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s.new", path.Dir(dst), content, dst))); err != nil {
		return err
	}
	if _, err := rnr.RunCmd(exec.Command("bash", "-c", fmt.Sprintf("sudo diff -u %s %s.new ; if [ $? != 0 ]; then sudo mv %s.new %s; sudo systemctl -f daemon-reload && sudo systemctl -f enable %s && sudo systemctl -f restart %s; fi", dst, dst, dst, dst, name, name))); err != nil {
		return err
	}
	return nil
}

// escapeSystemdDirectives escapes special characters in the input variables used to create the
// systemd unit file, which would otherwise be interpreted as systemd directives. An example
// are template specifiers (e.g. '%i') which are predefined variables that get evaluated dynamically
// (see systemd man pages for more info). This is not supported by minikube, thus needs to be escaped.
func escapeSystemdDirectives(engineConfigContext *engine.ConfigContext) {
	// escape '%' in Environment option so that it does not evaluate into a template specifier
	engineConfigContext.EngineOptions.Env = replaceChars(engineConfigContext.EngineOptions.Env, systemdSpecifierEscaper)
	// input might contain whitespaces, wrap it in quotes
	engineConfigContext.EngineOptions.Env = concatStrings(engineConfigContext.EngineOptions.Env, "\"", "\"")
}

// replaceChars returns a copy of the src slice with each string modified by the replacer
func replaceChars(src []string, replacer *strings.Replacer) []string {
	ret := make([]string, len(src))
	for i, s := range src {
		ret[i] = replacer.Replace(s)
	}
	return ret
}

// concatStrings concatenates each string in the src slice with prefix and postfix and returns a new slice
func concatStrings(src []string, prefix string, postfix string) []string {
	var buf bytes.Buffer
	ret := make([]string, len(src))
	for i, s := range src {
		buf.WriteString(prefix)
		buf.WriteString(s)
		buf.WriteString(postfix)
		ret[i] = buf.String()
		buf.Reset()
	}
	return ret
}
