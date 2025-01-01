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

package firewall

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// IsBootpdBlocked checks if the bootpd process is blocked by the macOS builtin firewall
func IsBootpdBlocked(cc config.ClusterConfig) bool {
	// only applies to qemu, on macOS, with socket_vmnet
	if cc.Driver != driver.QEMU2 || runtime.GOOS != "darwin" || cc.Network != "socket_vmnet" {
		return false
	}
	out, err := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").Output()
	if err != nil {
		klog.Warningf("failed to get firewall state: %v", err)
		return false
	}
	if regexp.MustCompile(`Firewall is disabled`).Match(out) {
		return false
	}
	out, err = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--listapps").Output()
	if err != nil {
		klog.Warningf("failed to list firewall apps: %v", err)
		return false
	}
	return !regexp.MustCompile(`\/usr\/libexec\/bootpd.*\n.*\( Allow`).Match(out)
}

// UnblockBootpd adds bootpd to the built-in macOS firewall and then unblocks it
func UnblockBootpd() error {
	cmds := []*exec.Cmd{
		exec.Command("sudo", "/usr/libexec/ApplicationFirewall/socketfilterfw", "--add", "/usr/libexec/bootpd"),
		exec.Command("sudo", "/usr/libexec/ApplicationFirewall/socketfilterfw", "--unblock", "/usr/libexec/bootpd"),
	}

	var cmdString strings.Builder
	for _, c := range cmds {
		cmdString.WriteString(fmt.Sprintf("    $ %s \n", strings.Join(c.Args, " ")))
	}

	out.Styled(style.Permissions, "Your firewall is blocking bootpd which is required for this configuration. The following commands will be executed to unblock bootpd:\n\n{{.commands}}\n", out.V{"commands": cmdString.String()})

	for _, c := range cmds {
		testArgs := append([]string{"-n"}, c.Args[1:]...)
		test := exec.Command("sudo", testArgs...)
		klog.Infof("testing: %s", test.Args)
		if err := test.Run(); err != nil {
			klog.Infof("%v may require a password: %v", c.Args, err)
			if !viper.GetBool("interactive") {
				klog.Warningf("%s requires a password, and --interactive=false", c.Args)
				c.Args = slices.Insert(c.Args, 1, "-n")
			}
		}
		klog.Infof("running: %s", c.Args)
		err := c.Run()
		if err != nil {
			return fmt.Errorf("running %s failed: %v", c.Args, err)
		}
	}
	return nil
}
