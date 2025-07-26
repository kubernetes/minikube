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

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.SeparateRows = true

	// Header
	var headers table.Row
	if cc == nil {
		headers = table.Row{"Addon Name", "Maintainer"}
	} else {
		headers = table.Row{"Addon Name", "Profile", "Status", "Maintainer"}
	}
	if printDocs {
		headers = append(headers, "Docs")
	}
	t.AppendHeader(headers)

	// Data
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

		if cc == nil {
			t.AppendRow(table.Row{addonName, maintainer})
			continue
		}

		enabled := addonBundle.IsEnabled(cc)
		statusStr := fmt.Sprintf("%s %s", stringFromStatus(enabled), iconFromStatus(enabled))

		row := table.Row{addonName, cc.Name, statusStr, maintainer}
		if printDocs {
			row = append(row, docs)
		}

		if enabled {
			// Color entire row green
			for i := range row {
				row[i] = text.FgGreen.Sprintf("%v", row[i])
			}
		}else {
			// Color entire row red
			for i := range row {
				row[i] = text.FgRed.Sprintf("%v", row[i])
			}
		}
		t.AppendRow(row)
	}

	t.Render()

	// Suggestion for multi-profile users
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
