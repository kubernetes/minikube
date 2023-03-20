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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type run struct {
	cmd    float64
	api    float64
	k8s    float64
	dnsSvc float64
	app    float64
	dnsAns float64
}

type runs struct {
	version string
	runs    []run
	cpus    []cpu
}

type cpu struct {
	cpuPct  float64 // percentage
	cpuTime float64 // second
}

func readInCSV(csvPath string, apps map[string]runs) error {
	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}

	r := csv.NewReader(f)
	for {
		d, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// skip the first line of the CSV file
		if d[0] == "name" {
			continue
		}

		values := []float64{}

		// 8-16 contain the run results
		for i := 8; i <= 16; i++ {
			v, err := strconv.ParseFloat(d[i], 64)
			if err != nil {
				return err
			}
			values = append(values, v)
		}
		newRun := run{values[0], values[1], values[2], values[3], values[4], values[5]}
		newCPU := cpu{values[6], values[8]}

		// get the app from the map and add the new run to it
		name := d[0]
		k, ok := apps[name]
		if !ok {
			k = runs{version: d[5]}
		}
		k.runs = append(k.runs, newRun)
		k.cpus = append(k.cpus, newCPU)
		apps[name] = k
	}

	return nil
}

func values(apps map[string]runs) ([]plotter.Values, []plotter.Values, []plotter.Values, []float64, []string) {
	var cmdValues, apiValues, k8sValues, dnsSvcValues, appValues, dnsAnsValues plotter.Values
	var cpuPctValues, cpuTimeValues plotter.Values
	var cpuMinikube, cpuKind, cpuK3d plotter.Values

	names := []string{}
	totals := []float64{}

	// for each app, calculate the average for all the runs, and append them to the charting values
	for _, name := range []string{"minikube", "kind", "k3d"} {
		app := apps[name]
		var cmd, api, k8s, dnsSvc, appRun, dnsAns float64
		var cpuPct, cpuTime float64

		names = append(names, app.version)

		for _, l := range app.runs {
			cmd += l.cmd
			api += l.api
			k8s += l.k8s
			dnsSvc += l.dnsSvc
			appRun += l.app
			dnsAns += l.dnsAns
		}

		for _, l := range app.cpus {
			cpuPct += l.cpuPct
			cpuTime += l.cpuTime
		}

		c := float64(len(app.runs))

		cmdAvg := cmd / c
		apiAvg := api / c
		k8sAvg := k8s / c
		dnsSvcAvg := dnsSvc / c
		appAvg := appRun / c
		dnsAnsAvg := dnsAns / c

		cpuPctAvg := cpuPct / c
		cpuTimeAvg := cpuTime / c

		cmdValues = append(cmdValues, cmdAvg)
		apiValues = append(apiValues, apiAvg)
		k8sValues = append(k8sValues, k8sAvg)
		dnsSvcValues = append(dnsSvcValues, dnsSvcAvg)
		appValues = append(appValues, appAvg)
		dnsAnsValues = append(dnsAnsValues, dnsAnsAvg)

		total := cmdAvg + apiAvg + k8sAvg + dnsSvcAvg + appAvg + dnsAnsAvg
		totals = append(totals, total)

		cpuPctValues = append(cpuPctValues, cpuPctAvg)
		cpuTimeValues = append(cpuTimeValues, cpuTimeAvg)

		cpuSummary := []float64{cpuPctAvg, cpuTimeAvg}

		switch name {
		case "minikube":
			cpuMinikube = cpuSummary
		case "kind":
			cpuKind = cpuSummary
		case "k3d":
			cpuK3d = cpuSummary
		}

	}

	runningTime := []plotter.Values{cmdValues, apiValues, k8sValues, dnsSvcValues, appValues, dnsAnsValues}

	// for markdown table, row is either cpu utilization or cpu time, col is process name
	cpu := []plotter.Values{cpuPctValues, cpuTimeValues}

	// row is process name, col is either cpu utilization, or cpu time
	cpureverse := []plotter.Values{cpuMinikube, cpuKind, cpuK3d}

	return runningTime, cpu, cpureverse, totals, names
}

func outputMarkdownTable(categories []plotter.Values, totals []float64, names []string) {
	headers := append([]string{""}, names...)
	c := [][]string{}
	fields := []string{"Command Exec", "API Server Answering", "Kubernetes SVC", "DNS SVC", "App Running", "DNS Answering"}
	for i, values := range categories {
		row := []string{fields[i]}
		for _, value := range values {
			row = append(row, fmt.Sprintf("%.3f", value))
		}
		c = append(c, row)
	}
	totalStrings := []string{"Total"}
	for _, t := range totals {
		totalStrings = append(totalStrings, fmt.Sprintf("%.3f", t))
	}
	c = append(c, totalStrings)
	b := new(bytes.Buffer)
	t := tablewriter.NewWriter(b)
	t.SetAutoWrapText(false)
	t.SetHeader(headers)
	t.SetAutoFormatHeaders(false)
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	t.AppendBulk(c)
	t.Render()
	data.TimeMarkdown = b.String()
}

func createChart(chartPath string, values []plotter.Values, totals []float64, names []string) error {
	p := plot.New()
	p.Title.Text = "Time to go from 0 to successful Kubernetes deployment"
	p.Y.Label.Text = "time (seconds)"

	bars := []*plotter.BarChart{}

	// create bars for all the values
	for i, v := range values {
		bar, err := createBars(v, i)
		if err != nil {
			return err
		}
		bars = append(bars, bar)
		p.Add(bar)
	}

	// stack the bars
	bars[0].StackOn(bars[1])
	bars[1].StackOn(bars[2])
	bars[2].StackOn(bars[3])
	bars[3].StackOn(bars[4])
	bars[4].StackOn(bars[5])

	// max Y value of the chart
	p.Y.Max = 80

	// add all the bars to the legend
	legends := []string{"Command Exec", "API Server Answering", "Kubernetes SVC", "DNS SVC", "App Running", "DNS Answering"}
	for i, bar := range bars {
		p.Legend.Add(legends[i], bar)
	}

	p.Legend.Top = true

	// add app name to the bars
	p.NominalX(names...)

	// create total time labels
	var labels []string
	for _, total := range totals {
		label := fmt.Sprintf("%.2f", total)
		labels = append(labels, label)
	}

	// create label positions
	var labelPositions []plotter.XY
	for i := range totals {
		x := float64(i) - 0.03
		y := totals[i] + 0.3
		labelPosition := plotter.XY{X: x, Y: y}
		labelPositions = append(labelPositions, labelPosition)
	}

	l, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    labelPositions,
		Labels: labels,
	},
	)
	if err != nil {
		return err
	}

	p.Add(l)

	return p.Save(12*vg.Inch, 8*vg.Inch, chartPath)
}

func createBars(values plotter.Values, index int) (*plotter.BarChart, error) {
	bars, err := plotter.NewBarChart(values, vg.Points(20))
	if err != nil {
		return nil, err
	}
	bars.LineStyle.Width = vg.Length(0)
	bars.Width = vg.Length(80)
	bars.Color = plotutil.Color(index)

	return bars, nil
}
