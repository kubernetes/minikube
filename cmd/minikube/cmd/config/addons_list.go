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
	"github.com/fatih/color"
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
	addonNames := slices.Sorted(maps.Keys(assets.Addons))
	table := tablewriter.NewWriter(os.Stdout)
	table.Options(tablewriter.WithHeaderAutoFormat(tw.On))

	// Table header
	header := []string{"Addon Name", "Maintainer"}
	if cc != nil {
		header = []string{"Addon Name", "Enabled", "Maintainer"}
	}
	if printDocs {
		header = append(header, "Docs")
	}
	table.Header(header)

	var rows [][]string

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

		// Step 1: build row
		var row []string
		if cc == nil {
			row = []string{addonName, maintainer}
		} else {
			enabled := addonBundle.IsEnabled(cc)
			status := iconFromStatus(enabled)
			row = []string{addonName, status, maintainer}
			if printDocs {
				row = append(row, docs)
			}

			// Step 2: apply coloring
			switch {
			case enabled:
				ColorRow(row, color.GreenString)
			default:
				ColorRow(row, color.WhiteString)
			}
		}

		if cc == nil && printDocs {
			row = append(row, docs)
		}

		rows = append(rows, row)
	}

	if err := table.Bulk(rows); err != nil {
		klog.Error("Error rendering table (bulk)", err)
	}
	if err := table.Render(); err != nil {
		klog.Error("Error rendering table", err)
	}

	// Profiles hint
	if v, _, err := config.ListProfiles(); err == nil && len(v) > 1 {
		out.Styled(style.Tip, "To see addons list for other profiles use: `minikube addons -p name list`")
	}
}

// ----------------
// Helpers
// ----------------

// colorFunc allows generic coloring
type colorFunc func(string, ...interface{}) string

func isEmoji(s string) bool {
	return strings.Contains(s, "✅")
}

func ColorRow(row []string, colored colorFunc) {
	for i := range row {
		if row[i] == "" || isEmoji(row[i]) {
			continue
		}
		row[i] = colored("%s", row[i])
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
