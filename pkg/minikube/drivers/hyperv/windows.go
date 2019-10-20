// +build windows

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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
	"errors"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/exit"
	"os/exec"
	"strings"
)

var powershell string

var (
	ErrPowerShellNotFound = errors.New("Powershell was not found in the path")
	ErrNotAdministrator   = errors.New("Hyper-v commands have to be run as an Administrator")
	ErrNotInstalled       = errors.New("Hyper-V PowerShell Module is not available")
)

func init() {
	var err error
	powershell, err = exec.LookPath("powershell.exe")
	if err != nil {
		exit.WithError("%v", ErrPowerShellNotFound)
	}
}

func powershellCmdOut(args ...string) (string, error) {
	args = append([]string{"-NoProfile", "-NonInteractive"}, args...)
	cmd := exec.Command(powershell, args...)
	glog.Infof("[executing ==>] : %v %v", powershell, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	glog.Infof("[stdout =====>] : %s", stdout.String())
	glog.Infof("[stderr =====>] : %s", stderr.String())
	return stdout.String(), err
}

func powershellCmd(args ...string) error {
	_, err := powershellCmdOut(args...)
	return err
}

func parseLines(stdout string) []string {
	resp := []string{}

	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		resp = append(resp, s.Text())
	}

	return resp
}

func hypervAvailable() error {
	stdout, err := powershellCmdOut("@(Get-Module -ListAvailable hyper-v).Name | Get-Unique")
	if err != nil {
		return err
	}

	resp := parseLines(stdout)
	if resp[0] != "Hyper-V" {
		return ErrNotInstalled
	}

	return nil
}

func IsAdministrator() (bool, error) {
	hypervAdmin := IsHypervAdministrator()

	if hypervAdmin {
		return true, nil
	}

	windowsAdmin, err := IsWindowsAdministrator()

	if err != nil {
		return false, err
	}

	return windowsAdmin, nil
}

func IsHypervAdministrator() bool {
	stdout, err := powershellCmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole("S-1-5-32-578")`)
	if err != nil {
		glog.Errorf("windows administrator check - %v", err)
		return false
	}

	resp := parseLines(stdout)
	return resp[0] == "True"
}

func IsWindowsAdministrator() (bool, error) {
	stdout, err := powershellCmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")`)
	if err != nil {
		return false, err
	}

	resp := parseLines(stdout)
	return resp[0] == "True", nil
}

// This function is used to get the full name of the current user trying to run. It will be DOMAIN\Username or MACHINE_NAME\Username
// TODO - Check if CIFS shares can be used by people who have domain accounts and are local admins on their machines.
func GetCurrentWindowsUser() (string, error) {
	stdout, err := powershellCmdOut(`@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).Identities.Name`)
	if err != nil {
		return "", err
	}
	response := parseLines(stdout)
	return response[0], nil
}