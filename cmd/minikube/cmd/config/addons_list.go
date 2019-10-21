/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
)

const defaultAddonListFormat = "- {{.AddonName}}: {{.AddonStatus}}\n"

var addonListFormat string
var addonListOutput string

// AddonListTemplate represents the addon list template
type AddonListTemplate struct {
	AddonName   string
	AddonStatus string
}

var addonsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all available minikube addons as well as their current statuses (enabled/disabled)",
	Long:  "Lists all available minikube addons as well as their current statuses (enabled/disabled)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			exit.UsageT("usage: minikube addons list")
		}

		if addonListOutput != "list" && addonListFormat != defaultAddonListFormat {
			exit.UsageT("Cannot use both --output and --format options")
		}

		switch strings.ToLower(addonListOutput) {
		case "list":
			printAddonsList()
		case "json":
			printAddonsJSON()
		default:
			exit.WithCodeT(exit.BadUsage, fmt.Sprintf("invalid output format: %s. Valid values: 'list', 'json'", addonListOutput))
		}
	},
}

func init() {
	addonsListCmd.Flags().StringVarP(
		&addonListFormat,
		"format",
		"f",
		defaultAddonListFormat,
		`Go template format string for the addon list output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list of accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd/config#AddonListTemplate`)

	addonsListCmd.Flags().StringVarP(
		&addonListOutput,
		"output",
		"o",
		"list",
		`minikube addons list --output OUTPUT. json, list`)

	AddonsCmd.AddCommand(addonsListCmd)
}

var stringFromStatus = func(addonStatus bool) string {
	if addonStatus {
		return "enabled"
	}
	return "disabled"
}

var printAddonsList = func() {
	addonNames := make([]string, 0, len(assets.Addons))
	for addonName := range assets.Addons {
		addonNames = append(addonNames, addonName)
	}
	sort.Strings(addonNames)

	for _, addonName := range addonNames {
		addonBundle := assets.Addons[addonName]
		addonStatus, err := addonBundle.IsEnabled()
		if err != nil {
			exit.WithError("Error getting addons status", err)
		}
		tmpl, err := template.New("list").Parse(addonListFormat)
		if err != nil {
			exit.WithError("Error creating list template", err)
		}
		listTmplt := AddonListTemplate{addonName, stringFromStatus(addonStatus)}
		err = tmpl.Execute(os.Stdout, listTmplt)
		if err != nil {
			exit.WithError("Error executing list template", err)
		}
	}
}

var printAddonsJSON = func() {
	addonNames := make([]string, 0, len(assets.Addons))
	for addonName := range assets.Addons {
		addonNames = append(addonNames, addonName)
	}
	sort.Strings(addonNames)

	addonsMap := map[string]map[string]interface{}{}

	for _, addonName := range addonNames {
		addonBundle := assets.Addons[addonName]

		addonStatus, err := addonBundle.IsEnabled()
		if err != nil {
			exit.WithError("Error getting addons status", err)
		}

		addonsMap[addonName] = map[string]interface{}{
			"Status": stringFromStatus(addonStatus),
		}
	}
	jsonString, _ := json.Marshal(addonsMap)

	out.String(string(jsonString))
}
