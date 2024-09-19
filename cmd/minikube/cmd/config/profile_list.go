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

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"

	"github.com/docker/machine/libmachine"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
)

var profileOutput string
var isLight bool

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all minikube profiles.",
	Long:  "Lists all valid minikube profiles and detects all possible invalid profiles.",
	Run: func(_ *cobra.Command, _ []string) {
		output := strings.ToLower(profileOutput)
		out.SetJSON(output == "json")
		go notify.MaybePrintUpdateTextFromGithub()

		switch output {
		case "json":
			printProfilesJSON()
		case "table":
			printProfilesTable()
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'table', 'json'", profileOutput))
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
		exit.Message(reason.UsageNoProfileRunning, "No minikube profile was found.")
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
	cps := config.ControlPlanes(*p.Config)
	if len(cps) == 0 {
		exit.Message(reason.GuestCpConfig, "No control-plane nodes found.")
	}
	statuses, err := cluster.GetStatus(api, p.Config)
	if err != nil {
		klog.Errorf("error getting statuses: %v", err)
		return "Unknown"
	}
	clusterStatus := cluster.GetState(statuses, ClusterFlagValue(), p.Config)

	return clusterStatus.StatusName
}

func renderProfilesTable(ps [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Profile", "VM Driver", "Runtime", "IP", "Port", "Version", "Status", "Nodes", "Active Profile", "Active Kubecontext"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.AppendBulk(ps)
	table.Render()
}

func profilesToTableData(profiles []*config.Profile) [][]string {
	var data [][]string
	currentProfile := ClusterFlagValue()
	for _, p := range profiles {
		cpIP := p.Config.KubernetesConfig.APIServerHAVIP
		cpPort := p.Config.APIServerPort
		if !config.IsHA(*p.Config) {
			cp, err := config.ControlPlane(*p.Config)
			if err != nil {
				exit.Error(reason.GuestCpConfig, "error getting control-plane node", err)
			}
			cpIP = cp.IP
			cpPort = cp.Port
		}

		k8sVersion := p.Config.KubernetesConfig.KubernetesVersion
		if k8sVersion == constants.NoKubernetesVersion { // for --no-kubernetes flag
			k8sVersion = "N/A"
		}
		var c, k string
		if p.Name == currentProfile {
			c = "*"
		}
		if p.ActiveKubeContext {
			k = "*"
		}
		data = append(data, []string{p.Name, p.Config.Driver, p.Config.KubernetesConfig.ContainerRuntime, cpIP, strconv.Itoa(cpPort), k8sVersion, p.Status, strconv.Itoa(len(p.Config.Nodes)), c, k})
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
		out.Errf("\t $ minikube delete -p %s \n", p.Name)
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
		os.Stdout.Write(jsonString)
	} else {
		body["error"] = err
		jsonString, _ := json.Marshal(body)
		os.Stdout.Write(jsonString)
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
	profileListCmd.Flags().StringVarP(&profileOutput, "output", "o", "table", "The output format. One of 'json', 'table'")
	profileListCmd.Flags().BoolVarP(&isLight, "light", "l", false, "If true, returns list of profiles faster by skipping validating the status of the cluster.")
	ProfileCmd.AddCommand(profileListCmd)
}
