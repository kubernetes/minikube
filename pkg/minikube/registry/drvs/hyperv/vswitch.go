//go:build windows

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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type netAdapter struct {
	InterfaceGUID        string `json:"interfaceGuid"`
	InterfaceDescription string
}

type vmSwitch struct {
	Name                    string
	NetAdapterInterfaceGuid []string
}

// returns network adapters matching to the given filtering condition
func getNetAdapters(physical bool, condition string) ([]netAdapter, error) {
	cmdlet := []string{"Get-NetAdapter"}
	if physical {
		cmdlet = append(cmdlet, "-Physical")
	}
	cmd := []string{strings.Join(cmdlet, " ")}
	if condition != "" {
		cmd = append(cmd, fmt.Sprintf("Where-Object {%s}", condition))
	}
	cmd = append(cmd, "Select-Object -Property InterfaceGuid, InterfaceDescription")
	stdout, err := cmdOut(fmt.Sprintf("ConvertTo-Json @(%s)", strings.Join(cmd, " | ")))
	if err != nil {
		return nil, err
	}

	var adapters []netAdapter
	err = json.Unmarshal([]byte(strings.TrimSpace(stdout)), &adapters)
	if err != nil {
		return nil, err
	}

	return adapters, nil
}

// returns Hyper-V switches matching to the given filtering condition
func getVMSwitch(condition string) ([]vmSwitch, error) {
	cmd := []string{"Hyper-V\\Get-VMSwitch"}
	if condition != "" {
		cmd = append(cmd, fmt.Sprintf("Where-Object {%s}", condition))
	}
	cmd = append(cmd, "Select-Object -Property Name, NetAdapterInterfaceGuid")
	stdout, err := cmdOut(fmt.Sprintf("ConvertTo-Json @(%s)", strings.Join(cmd, " | ")))
	if err != nil {
		return nil, err
	}

	var vmSwitches []vmSwitch
	err = json.Unmarshal([]byte(strings.TrimSpace(stdout)), &vmSwitches)
	if err != nil {
		return nil, err
	}

	return vmSwitches, nil
}

// returns Hyper-V switches which connects to the adapter of the given GUID
func findConnectedVMSwitch(adapterGUID string) (string, error) {
	foundSwitches, err := getVMSwitch(fmt.Sprintf("($_.SwitchType -eq 2) -And ($_.NetAdapterInterfaceGuid -contains \"%s\")", adapterGUID))
	if err != nil {
		return "", err
	}

	if len(foundSwitches) > 0 {
		return foundSwitches[0].Name, nil
	}

	return "", nil
}

// returns "up" net adapters in the order Physical LAN adapters then other adapters
func getOrderedAdapters() ([]netAdapter, error) {
	// look for "Up" adapters and prefer physical LAN over other options
	lanAdapters, err := getNetAdapters(true, "($_.Status -eq \"Up\") -And ($_.PhysicalMediaType -like \"*802.3*\")")
	if err != nil {
		return nil, err
	}

	upAdapters, err := getNetAdapters(true, "$_.Status -eq \"Up\"")
	if err != nil {
		return nil, err
	}

	var orderedAdapters []netAdapter
	adapterGuids := map[string]interface{}{}

	// all adapters will be checked in the following order:
	//  1. Connected physical LAN adapters
	//  2. Any other connected adapters
	for _, adapter := range lanAdapters {
		if _, ok := adapterGuids[adapter.InterfaceGUID]; !ok {
			adapterGuids[adapter.InterfaceGUID] = nil
			orderedAdapters = append(orderedAdapters, adapter)
		}
	}

	for _, adapter := range upAdapters {
		if _, ok := adapterGuids[adapter.InterfaceGUID]; !ok {
			adapterGuids[adapter.InterfaceGUID] = nil
			orderedAdapters = append(orderedAdapters, adapter)
		}
	}

	return orderedAdapters, nil
}

// create a new VM switch of the given name and network adapter
func createVMSwitch(switchName string, adapter netAdapter) error {
	err := cmd(fmt.Sprintf("Hyper-V\\New-VMSwitch -Name \"%s\" -NetAdapterInterfaceDescription \"%s\"", switchName, adapter.InterfaceDescription))
	if err != nil {
		return errors.Wrapf(err, "failed to create VM switch %s with adapter %s", switchName, adapter.InterfaceGUID)
	}

	return nil
}

// choose VM switch connected to an adapter. If adapter name is not specified,
// it tries to use an "up" LAN adapter then other adapters for external network
func chooseSwitch(adapterName string) (string, netAdapter, error) {
	var adapter netAdapter
	if adapterName != "" {
		foundAdapters, err := getNetAdapters(false, fmt.Sprintf("($_.InterfaceDescription -eq \"%s\")", adapterName))
		if err != nil {
			return "", netAdapter{}, err
		}

		if len(foundAdapters) == 0 {
			return "", netAdapter{}, errors.Errorf("adapter %s not found", adapterName)
		}

		adapter = foundAdapters[0]
		foundSwitch, err := findConnectedVMSwitch(adapter.InterfaceGUID)
		return foundSwitch, adapter, err
	}

	adapters, err := getOrderedAdapters()
	if err != nil {
		return "", netAdapter{}, err
	}

	if len(adapters) == 0 {
		return "", netAdapter{}, errors.Errorf("no connected adapter available")
	}

	externalVMSwitches, err := getVMSwitch("($_.SwitchType -eq 2)")
	if err != nil {
		return "", netAdapter{}, errors.Wrapf(err, "failed to list external VM switches")
	}

	if len(externalVMSwitches) > 0 {
		// it doesn't seem like Windows allows one VM switch for each adapter
		adapterSwitches := map[string][]string{}
		for _, vmSwitch := range externalVMSwitches {
			for _, connectedAdapter := range vmSwitch.NetAdapterInterfaceGuid {
				var switches []string
				key := strings.ToUpper(fmt.Sprintf("{%s}", connectedAdapter))
				if _, ok := adapterSwitches[key]; ok {
					switches = adapterSwitches[key]
				}
				switches = append(switches, vmSwitch.Name)
				adapterSwitches[key] = switches
			}
		}

		for _, adapter := range adapters {
			if switches, ok := adapterSwitches[adapter.InterfaceGUID]; ok && len(switches) > 0 {
				return switches[0], adapter, nil
			}
		}
	}

	adapter = adapters[0]
	return "", adapter, nil
}
