/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

var page = `---
title: "{{.Version}} Benchmark"
linkTitle: "{{.Version}} Benchmark"
weight: -{{.Weight}}
---

![time-to-k8s](/images/benchmarks/timeToK8s/{{.Version}}-time.png)

{{.TimeMarkdown}}


![cpu-to-k8s](/images/benchmarks/timeToK8s/{{.Version}}-cpu.png)

{{.CPUMarkdown}}
`

type Data struct {
	Version      string
	Weight       string
	TimeMarkdown string
	CPUMarkdown  string
}

var data Data

func main() {
	csvPath := flag.String("csv", "", "path to the CSV file")
	imagePath := flag.String("image", "", "path to output the chart to")
	pagePath := flag.String("page", "", "path to output the page to")

	flag.Parse()

	data.Weight = time.Now().Format("20060102")

	version, err := exec.Command("minikube", "version", "--short").Output()
	if err != nil {
		log.Fatalf("failed to get minikube version: %v", err)
	}
	data.Version = strings.Split(string(version), "\n")[0]

	// map of the apps (minikube, kind, k3d) and their runs
	apps := make(map[string]runs)

	if err := readInCSV(*csvPath, apps); err != nil {
		log.Fatalf("fail to readin cvs file with err %s", err)
	}

	runningTime, cpuMdPlot, cpuChartPlot, totals, names := values(apps)

	// markdown table for running time
	outputMarkdownTable(runningTime, totals, names)

	// chart for running time
	if err := createChart(*imagePath+"-time.png", runningTime, totals, names); err != nil {
		log.Fatalf("fail to create running time chart with err %s", err)
	}

	// markdown table for cpu utilization
	cpuMarkdownTable(cpuMdPlot, names)

	// chart for cpu utilization
	if err := createCPUChart(*imagePath+"-cpu.png", cpuChartPlot, names); err != nil {
		log.Fatalf("fail to create CPU chart with err %s", err)
	}

	// generate page and save
	tmpl, err := template.New("msg").Parse(page)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(*pagePath)
	if err != nil {
		log.Fatalf("fail to create file under path %s, with err %s", *pagePath, err)
	}

	if err = tmpl.Execute(f, data); err != nil {
		log.Fatalf("fail to populate the page with err %s", err)
	}

	f.Close()
}
