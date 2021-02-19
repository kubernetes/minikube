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

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
)

var output string
var isLight bool

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all minikube profiles.",
	Long:  "Lists all valid minikube profiles and detects all possible invalid profiles.",
	Run: func(cmd *cobra.Command, args []string) {
		switch strings.ToLower(output) {
		case "json":
			printProfilesJSON()
		case "table":
			printProfilesTable()
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'table', 'json'", output))
		}
	},
}

func listProfiles() (validProfiles, invalidProfiles []*config.Profile, err error) {
	if isLight {
		validProfiles, err = config.ListValidProfiles()
	} else {
		validProfiles, invalidProfiles, err = config.ListProfiles()
	}

	return validProfiles, invalidProfiles, err
}

func printProfilesTable() {
	validProfiles, invalidProfiles, err := listProfiles()

	if err != nil {
		klog.Warningf("error loading profiles: %v", err)
	}

	if len(validProfiles) == 0 {
		exit.Message(reason.Usage, "No minikube profile was found. You can create one using `minikube start`.")
	}

	updateProfilesStatus(validProfiles)
	renderProfilesTable(profilesToTableData(validProfiles))
	warnInvalidProfiles(invalidProfiles)
}

func updateProfilesStatus(profiles []*config.Profile) {
	if isLight {
		for _, p := range profiles {
			p.Status = "Skipped"
		}
		return
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		klog.Errorf("failed to get machine api client %v", err)
	}
	defer api.Close()

	for _, p := range profiles {
		p.Status = profileStatus(p, api)
	}
}

func profileStatus(p *config.Profile, api libmachine.API) string {
	cp, err := config.PrimaryControlPlane(p.Config)
	if err != nil {
		exit.Error(reason.GuestCpConfig, "error getting primary control plane", err)
	}

	host, err := machine.LoadHost(api, config.MachineName(*p.Config, cp))
	if err != nil {
		klog.Warningf("error loading profiles: %v", err)
		return "Unknown"
	}

	// The machine isn't running, no need to check inside
	s, err := host.Driver.GetState()
	if err != nil {
		klog.Warningf("error getting host state: %v", err)
		return "Unknown"
	}
	if s != state.Running {
		return s.String()
	}

	cr, err := machine.CommandRunner(host)
	if err != nil {
		klog.Warningf("error loading profiles: %v", err)
		return "Unknown"
	}

	hostname, _, port, err := driver.ControlPlaneEndpoint(p.Config, &cp, host.DriverName)
	if err != nil {
		klog.Warningf("error loading profiles: %v", err)
		return "Unknown"
	}

	status, err := kverify.APIServerStatus(cr, hostname, port)
	if err != nil {
		klog.Warningf("error getting apiserver status for %s: %v", p.Name, err)
		return "Unknown"
	}
	return status.String()
}

func renderProfilesTable(ps [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Profile", "VM Driver", "Runtime", "IP", "Port", "Version", "Status", "Nodes"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.AppendBulk(ps)
	table.Render()
}

func profilesToTableData(profiles []*config.Profile) [][]string {
	var data [][]string
	for _, p := range profiles {
		cp, err := config.PrimaryControlPlane(p.Config)
		if err != nil {
			exit.Error(reason.GuestCpConfig, "error getting primary control plane", err)
		}

		data = append(data, []string{p.Name, p.Config.Driver, p.Config.KubernetesConfig.ContainerRuntime, cp.IP, strconv.Itoa(cp.Port), p.Config.KubernetesConfig.KubernetesVersion, p.Status, strconv.Itoa(len(p.Config.Nodes))})
	}
	return data
}

func warnInvalidProfiles(invalidProfiles []*config.Profile) {
	if invalidProfiles == nil {
		return
	}

	out.WarningT("Found {{.number}} invalid profile(s) ! ", out.V{"number": len(invalidProfiles)})
	for _, p := range invalidProfiles {
		out.ErrT(style.Empty, "\t "+p.Name)
	}

	out.ErrT(style.Tip, "You can delete them using the following command(s): ")
	for _, p := range invalidProfiles {
		out.Err(fmt.Sprintf("\t $ minikube delete -p %s \n", p.Name))
	}
}

func printProfilesJSON() {
	validProfiles, invalidProfiles, err := listProfiles()

	updateProfilesStatus(validProfiles)

	var body = map[string]interface{}{}
	if err == nil || config.IsNotExist(err) {
		body["valid"] = profilesOrDefault(validProfiles)
		body["invalid"] = profilesOrDefault(invalidProfiles)
		jsonString, _ := json.Marshal(body)
		out.String(string(jsonString))
	} else {
		body["error"] = err
		jsonString, _ := json.Marshal(body)
		out.String(string(jsonString))
		os.Exit(reason.ExGuestError)
	}
}

func profilesOrDefault(profiles []*config.Profile) []*config.Profile {
	if profiles != nil {
		return profiles
	}
	return []*config.Profile{}
}

func init() {
	profileListCmd.Flags().StringVarP(&output, "output", "o", "table", "The output format. One of 'json', 'table'")
	profileListCmd.Flags().BoolVarP(&isLight, "light", "l", false, "If true, returns list of profiles faster by skipping validating the status of the cluster.")
	ProfileCmd.AddCommand(profileListCmd)
}
