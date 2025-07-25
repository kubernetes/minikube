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
    "github.com/fatih/color"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
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

	// Prepare table header
	var tHeader []string
	var numCols int
	if cc == nil {
		tHeader = []string{"Addon Name", "Maintainer"}
		numCols = 2
	} else {
		tHeader = []string{"Addon Name", "Profile", "Status", "Maintainer"}
		numCols = 4
	}
	if printDocs {
		tHeader = append(tHeader, "Docs")
		numCols++
	}

	// Define color configuration
	colorCfg := renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG: renderer.Colors{color.FgHiGreen, color.Bold},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.FgCyan},
			Columns: []renderer.Tint{
				{FG: renderer.Colors{color.FgMagenta}}, // column 0
				{},                                     // column 1 (default)
				{FG: renderer.Colors{color.FgHiRed}},   // column 2
			},
		},
		Border:    renderer.Tint{FG: renderer.Colors{color.FgWhite}},
		Separator: renderer.Tint{FG: renderer.Colors{color.FgWhite}},
	}

	// Table config
	cfg := tablewriter.Config{
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{AutoWrap: tw.WrapNormal},
			Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
		},
		Footer: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignRight},
		},
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
		tablewriter.WithConfig(cfg),
	)

	table.Header(tHeader)

	// Construct table rows
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

		var row []string
		if cc == nil {
			row = []string{addonName, maintainer}
		} else {
			isEnabled := addonBundle.IsEnabled(cc)
			statusStr := fmt.Sprintf("%s %s", stringFromStatus(isEnabled), iconFromStatus(isEnabled))
			row = []string{addonName, cc.Name, statusStr, maintainer}
		}
		if printDocs {
			row = append(row, docs)
		}
		rows = append(rows, row)
	}

	table.Bulk(rows)

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
