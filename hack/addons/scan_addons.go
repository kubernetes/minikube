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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

func main() {
	addonsFile, err := os.ReadFile("hack/addons/addons.yaml")
	if err != nil {
		panic(err)
	}
	addons := make(map[string]interface{})
	err = yaml.Unmarshal(addonsFile, addons)
	if err != nil {
		panic(err)
	}

	for a, i := range addons {
		fmt.Printf("ADDON: %s\n", a)
		images := i.(map[interface{}]interface{})
		for _, image := range images {
			fmt.Println(image)
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
			vulnz := outmap["vulnerabilities"].([]interface{})
			for _, v := range vulnz {
				vuln := v.(map[string]interface{})
				fmt.Printf("%s, %s, %s, %s\n", vuln["title"], vuln["packageName"], vuln["severity"], vuln["nearestFixedInVersion"])
			}
		}
	}
}
