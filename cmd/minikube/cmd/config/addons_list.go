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
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/cmd/minikube/cmd/flags"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var addonListOutput string
var addonPrintDocs bool

// AddonListTemplate represents the addon list template
type AddonListTemplate struct {
	AddonName   string
	AddonStatus string
}

var addonsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all available minikube addons as well as their current statuses (enabled/disabled)",
	Long:  "Lists all available minikube addons as well as their current statuses (enabled/disabled)",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) != 0 {
			exit.Message(reason.Usage, "usage: minikube addons list")
		}

		options := flags.CommandOptions()
		var cc *config.ClusterConfig
		if config.ProfileExists(ClusterFlagValue()) {
			_, cc = mustload.Partial(ClusterFlagValue(), options)
		}
		switch strings.ToLower(addonListOutput) {
		case "list":
			printAddonsList(cc, addonPrintDocs)
		case "json":
			printAddonsJSON(cc)
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'list', 'json'", addonListOutput))
		}
	},
}

func init() {
	addonsListCmd.Flags().StringVarP(&addonListOutput, "output", "o", "list", "minikube addons list --output OUTPUT. json, list")
	addonsListCmd.Flags().BoolVarP(&addonPrintDocs, "docs", "d", false, "If true, print web links to addons' documentation if using --output=list (default).")
	AddonsCmd.AddCommand(addonsListCmd)
}

var iconFromStatus = func(addonStatus bool) string {
	if addonStatus {
		return "✅"
	}
	return "   " // because emoji indentation is different
}

var stringFromStatus = func(addonStatus bool) string {
	if addonStatus {
		return "enabled"
	}
	return "disabled"
}

var printAddonsList = func(cc *config.ClusterConfig, printDocs bool) {
	addonNames := slices.Sorted(maps.Keys(assets.Addons))
	table := tablewriter.NewWriter(os.Stdout)

	table.Options(
		tablewriter.WithHeaderAutoFormat(tw.On),
	)

	// Create table header
	var tHeader []string
	if cc == nil {
		tHeader = []string{"Addon Name", "Maintainer"}
	} else {
		tHeader = []string{"Addon Name", "Profile", "Status", "Maintainer"}
	}
	if printDocs {
		tHeader = append(tHeader, "Docs")
	}
	table.Header(tHeader)

	// Create table data
	for _, addonName := range addonNames {
		addonBundle := assets.Addons[addonName]
		maintainer := addonBundle.Maintainer
		if maintainer == "" {
			maintainer = "3rd party (unknown)"
		}
		docs := addonBundle.Docs
		if docs == "" {
			docs = "n/a"
		}

		enabled := false
		if cc != nil {
			enabled = addonBundle.IsEnabled(cc)
		}

		// Prepare row data
		var row []string
		if cc == nil {
			row = []string{addonName, maintainer}
		} else {
			row = []string{addonName, cc.Name, fmt.Sprintf("%s %s", stringFromStatus(enabled), iconFromStatus(enabled)), maintainer}
		}

		if printDocs {
			row = append(row, docs)
		}

		// Apply green color if enabled
		if enabled {
			for i, val := range row {
				row[i] = style.Green + val + style.Reset
			}
		}

		table.Append(row)
	}
	if err := table.Render(); err != nil {
		klog.Error("Error rendering table", err)
	}
	v, _, err := config.ListProfiles()
	if err != nil {
		klog.Errorf("list profiles returned error: %v", err)
	}
	if len(v) > 1 {
		out.Styled(style.Tip, "To see addons list for other profiles use: `minikube addons -p name list`")
	}
}

var printAddonsJSON = func(cc *config.ClusterConfig) {
	addonNames := slices.Sorted(maps.Keys(assets.Addons))
	addonsMap := map[string]map[string]interface{}{}

	for _, addonName := range addonNames {
		if cc == nil {
			addonsMap[addonName] = map[string]interface{}{}
			continue
		}

		addonBundle := assets.Addons[addonName]
		enabled := addonBundle.IsEnabled(cc)
		addonsMap[addonName] = map[string]interface{}{
			"Status":  stringFromStatus(enabled),
			"Profile": cc.Name,
		}
		if addonPrintDocs {
			addonsMap[addonName]["Maintainer"] = addonBundle.Maintainer
			addonsMap[addonName]["Docs"] = addonBundle.Docs
		}
	}

	jsonString, _ := json.Marshal(addonsMap)
	out.String(string(jsonString))

}
