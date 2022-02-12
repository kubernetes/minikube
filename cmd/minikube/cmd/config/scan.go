/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/localpath"
)

var addonsScan = &cobra.Command{
	Use:    "scan",
	Short:  "Scans all minikube addon images for security vulnerabilites",
	Long:   "Scans all minikube addon images for security vulnerabilites",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := download.AddonList()
		if err != nil {
			panic(err)
		}
		addonsFile, err := os.ReadFile(filepath.Join(localpath.MiniPath(), "addons", "addons.yaml"))
		if err != nil {
			panic(err)
		}
		addonMap := make(map[string]interface{})
		err = yaml.Unmarshal(addonsFile, addonMap)
		if err != nil {
			panic(err)
		}

		addonStatus := make(map[string]addons.AddonStatus)

		for a, i := range addonMap {
			fmt.Printf("ADDON: %s\n", a)
			images := i.(map[interface{}]interface{})
			status := addons.AddonStatus{Enabled: true, Images: []addons.ImageStatus{}}
			for _, image := range images {
				fmt.Println(image)
				imageStatus := addons.ImageStatus{Image: image.(string), CVEs: []addons.CVE{}}
				snyk := exec.Command("snyk", "container", "test", image.(string), "--json", "--severity-threshold=high")
				out, err := snyk.Output()
				if err == nil {
					fmt.Println("no vulnerabilites found")
					continue
				}
				outmap := make(map[string]interface{})
				err = json.Unmarshal(out, &outmap)
				if err != nil {
					fmt.Printf("error unmarshalling json for %s: %v", image, err)
				}
				// The vulnerabilities entry won't show up if there was an error from snyk
				if vulnz, ok := outmap["vulnerabilities"].([]interface{}); ok {
					for _, v := range vulnz {
						vuln := v.(map[string]interface{})
						logCVE := true
						for _, c := range imageStatus.CVEs {
							if c.Name == vuln["title"].(string) {
								fmt.Printf("already logged CVE %s, skipping\n", c.Name)
								logCVE = false
								break
							}
						}
						if !logCVE {
							continue
						}
						fmt.Printf("%s, %s, %s, %s\n", vuln["title"], vuln["packageName"], vuln["severity"], vuln["nearestFixedInVersion"])
						status.Enabled = false
						updatedVersion := ""
						if uv, ok := vuln["nearestFixedInVersion"].(string); ok {
							updatedVersion = uv
						}
						cve := addons.CVE{
							Name:           vuln["title"].(string),
							PackageName:    vuln["packageName"].(string),
							Severity:       vuln["severity"].(string),
							UpdatedVersion: updatedVersion,
						}
						imageStatus.CVEs = append(imageStatus.CVEs, cve)
					}
				}
				if len(imageStatus.CVEs) > 0 {
					status.Images = append(status.Images, imageStatus)
				}
			}
			addonStatus[a] = status
		}
		statusYaml, err := yaml.Marshal(addonStatus)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("hack/addons/status.yaml", statusYaml, 0777)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	AddonsCmd.AddCommand(addonsScan)
}
