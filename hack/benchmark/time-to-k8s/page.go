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
	"fmt"
	"log"
	"os"
	"text/template"
	"time"
)

var page = `---
title: "{{.Version}} Benchmark"
linkTitle: "{{.Version}} Benchmark"
weight: -{{.Weight}}
---

![time-to-k8s]({{.TimeChart}})

{{.TimeMarkdown}}


![cpu-to-k8s]({{.CPUChart}})

{{.CPUMarkdown}}
`

type Data struct {
	Version      string
	Weight       string
	TimeChart    string
	TimeMarkdown string
	CPUChart     string
	CPUMarkdown  string
}

var data Data

func main() {
	csvPath := flag.String("csv", "", "path to the CSV file")
	imagePath := flag.String("image", "", "path to output the chart to")
	pagePath := flag.String("page", "", "path to output the page to")

	flag.Parse()

	t := time.Now()
	data.Weight = fmt.Sprintf("%d%d%d", t.Year(), t.Month(), t.Day())

	// map of the apps (minikube, kind, k3d) and their runs
	apps := make(map[string]runs)

	if err := readInCSV(*csvPath, apps); err != nil {
		log.Fatal(err)
	}

	runningTime, cpuMdPlot, cpuChartPlot, totals, names := values(apps)

	// markdown table for running time
	outputMarkdownTable(runningTime, totals, names)

	// chart for running time
	if err := createChart(*imagePath+"-time.png", runningTime, totals, names); err != nil {
		log.Fatal(err)
	}

	// markdown table for cpu utilization
	cpuMarkdownTable(cpuMdPlot, names)

	// chart for cpu utilization
	if err := createCPUChart(*imagePath+"-cpu.png", cpuChartPlot, names); err != nil {
		log.Fatal(err)
	}

	// generate page and save
	tmpl, err := template.New("msg").Parse(page)

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(*pagePath)
	if err != nil {
		log.Fatal(err)
	}

	if err = tmpl.Execute(f, data); err != nil {
		log.Fatal(err)
	}

	f.Close()

}
