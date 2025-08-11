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

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/constants"
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

		var cc *config.ClusterConfig
		if config.ProfileExists(ClusterFlagValue()) {
			_, cc = mustload.Partial(ClusterFlagValue())
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
	addonNames := make([]string, 0, len(assets.Addons))
	for addonName := range assets.Addons {
		addonNames = append(addonNames, addonName)
	}
	sort.Strings(addonNames)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(true)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")

	// Create table header
	var tHeader []string
	if cc == nil {
		tHeader = []string{"Addon Name", "Maintainer"}
	} else {
		tHeader = []string{"Addon Name", "Enabled", "Maintainer"}
	}
	if printDocs {
		tHeader = append(tHeader, "Docs")
	}
	table.SetHeader(tHeader)

	// Create table data
	var tData [][]string
	var temp []string
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
		
		// Determine base row structure and colors
		var enabled bool
		var colorCode string
		var applyColor bool
		
		if cc != nil {
			enabled = addonBundle.IsEnabled(cc)
			if enabled {
				colorCode = constants.Enabled
				applyColor = true
			}
		}
		
		// Build base row data
		var rawValues []interface{}
		if cc == nil {
			rawValues = []interface{}{addonName, maintainer}
		} else {
			status := ""
			if enabled {
				status = iconFromStatus(enabled)
			}
			rawValues = []interface{}{addonName, status, maintainer}
		}
		
		// Add docs if needed
		if printDocs {
			rawValues = append(rawValues, docs)
		}
		
		// Apply colors using loop
		temp = make([]string, len(rawValues))
		for i, value := range rawValues {
			valueStr := fmt.Sprintf("%v", value)
			
			// Apply color based on context
			shouldColorValue := false
			var valueColorCode string
			
			if cc == nil {
				// No coloring for null cluster config
				temp[i] = valueStr
				continue
			}
			
			switch i {
			case 0: // addonName
				shouldColorValue = applyColor
				valueColorCode = colorCode
			case 1: // status (only for non-null cc)
				shouldColorValue = applyColor
				valueColorCode = colorCode
			case 2: // maintainer
				shouldColorValue = applyColor
				valueColorCode = colorCode
			default: // docs or other columns
				if printDocs && i == len(rawValues)-1 {
					// This is the docs column
					if enabled {
						shouldColorValue = true
						valueColorCode = constants.Enabled
					} else {
						shouldColorValue = true
						valueColorCode = constants.Disabled
					}
				}
			}
			
			if shouldColorValue {
				temp[i] = fmt.Sprintf("%s%s%s", valueColorCode, valueStr, constants.Default)
			} else {
				temp[i] = valueStr
			}
		}
		tData = append(tData, temp)
	}
	table.AppendBulk(tData)

	table.Render()

	v, _, err := config.ListProfiles()
	if err != nil {
		klog.Errorf("list profiles returned error: %v", err)
	}
	if len(v) > 1 {
		out.Styled(style.Tip, "To see addons list for other profiles use: `minikube addons -p name list`")
	}
}

var printAddonsJSON = func(cc *config.ClusterConfig) {
	addonNames := make([]string, 0, len(assets.Addons))
	for addonName := range assets.Addons {
		addonNames = append(addonNames, addonName)
	}
	sort.Strings(addonNames)

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
