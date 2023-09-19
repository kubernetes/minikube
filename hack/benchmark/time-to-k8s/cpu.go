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
	"fmt"

	"github.com/olekukonko/tablewriter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var fields = []string{"CPU Utilization(%)", "CPU Time(seconds)"}

func cpuMarkdownTable(categories []plotter.Values, names []string) {

	// categories row is the either cpu pct or time, col is process name
	headers := append([]string{""}, names...)
	c := [][]string{}
	for i, values := range categories {
		row := []string{fields[i]}
		for _, value := range values {
			row = append(row, fmt.Sprintf("%.3f", value))
		}
		c = append(c, row)
	}
	b := new(bytes.Buffer)
	t := tablewriter.NewWriter(b)
	t.SetAutoWrapText(false)
	t.SetHeader(headers)
	t.SetAutoFormatHeaders(false)
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	t.AppendBulk(c)
	t.Render()
	data.CPUMarkdown = b.String()
}

func createCPUChart(chartPath string, values []plotter.Values, names []string) error {
	p := plot.New()
	p.Title.Text = "CPU utilization to go from 0 to successful Kubernetes deployment"
	p.Y.Label.Text = "CPU utilization"
	setYMax(p, values)

	w := vg.Points(20)

	barsA, err := plotter.NewBarChart(values[0], w)
	if err != nil {
		return err
	}
	barsA.LineStyle.Width = vg.Length(0)
	barsA.Color = plotutil.Color(0)
	barsA.Offset = -w

	barsB, err := plotter.NewBarChart(values[1], w)
	if err != nil {
		return err
	}
	barsB.LineStyle.Width = vg.Length(0)
	barsB.Color = plotutil.Color(1)
	barsB.Offset = 0

	barsC, err := plotter.NewBarChart(values[2], w)
	if err != nil {
		return err
	}
	barsC.LineStyle.Width = vg.Length(0)
	barsC.Color = plotutil.Color(2)
	barsC.Offset = w

	p.Add(barsA, barsB, barsC)
	p.Legend.Add(names[0], barsA)
	p.Legend.Add(names[1], barsB)
	p.Legend.Add(names[2], barsC)

	p.Legend.Top = true
	p.NominalX(fields...)

	return p.Save(8*vg.Inch, 8*vg.Inch, chartPath)
}

func setYMax(p *plot.Plot, values []plotter.Values) {
	ymax := 0.0
	for _, value := range values {
		for _, v := range value {
			if v > ymax {
				ymax = v
			}
		}
	}

	p.Y.Max = ymax + 5
}
