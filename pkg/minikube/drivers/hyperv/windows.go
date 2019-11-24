// +build windows

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package hyperv

import (
	"bufio"
	"bytes"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

var powershell string

var (
	ErrPowerShellNotFound 			= 	errors.New("Powershell was not found in the path")
	ErrNotWindowsAdministrator   	= 	errors.New("Command has to run as Administrator")
	ErrNotHyperVAdministrator		=	errors.New("Hyper-V Commands need to be run either as an Administrator or as a member of the group of Hyper-V Administrators")
)

func PowershellCmdOut(args ...string) (string, error) {

	// Check if PowerShell is available or not.
	var err error
	powershell, err = exec.LookPath("powershell.exe")
	if err != nil {
		return "", ErrPowerShellNotFound
	}

	args = append([]string{"-NoProfile", "-NonInteractive"}, args...)
	cmd := exec.Command(powershell, args...)
	glog.Infof("[executing] : %v", cmd.Args)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	glog.Infof("[stdout =====>] : %s", stdout.String())
	glog.Infof("[stderr =====>] : %s", stderr.String())
	return stdout.String(), err
}

// The first argument which is passed to this function must be the fully qualified module followed by the command.
// For example - "SmbShare\\New-SmbShare" or "Hyper-V\\New-VM"
func PowershellCmd(args ...string) error {
	_, err := PowershellCmdOut(args...)
	return err
}

func ParseLines(stdout string) []string {
	var resp []string

	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		resp = append(resp, s.Text())
	}

	return resp
}

func IsHyperVAdministrator() error {
	stdout, err := PowershellCmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole("S-1-5-32-578")`)
	if err != nil {
		return errors.Wrap(err,"hyper-v administrator")
	}

	resp := ParseLines(stdout)
	if resp[0] == "True" {
		return nil
	} else {
		return ErrNotHyperVAdministrator
	}
}

func IsWindowsAdministrator() error {
	stdout, err := PowershellCmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")`)
	if err != nil {
		return errors.Wrap(err,"current user role")
	}

	resp := ParseLines(stdout)
	if resp[0] == "True" {
		return nil
	} else {
		return ErrNotWindowsAdministrator
	}
}

/* This function is used to get the full name of the current user trying to run. It will be DOMAIN\Username or MACHINE_NAME\Username
	The function which comes with Golang i.e. user.Current() returns multiple things and need to be parsed multiple times. For example,
	it returned this string -- [&{S-1-5-21-3668668534-1506658371-4068906332-1001 S-1-5-21-3668668534-1506658371-4068906332-1001 DESKTOP\blue John Doe C:\Users\blue}]
	We only need the machine name and the user name. It returns the User SID, Machine Name, Username, User Complete Name, User Profile Path

	This function returns MachineName/Username which is what we require.
*/
// TODO - Check if CIFS shares can be used by people who have domain accounts and are local admins on their machines.
func CurrentWindowsUser() (string, error) {
	stdout, err := PowershellCmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).Identities.Name`)
	if err != nil {
		return "", err
	}
	response := ParseLines(stdout)
	return response[0], nil
}