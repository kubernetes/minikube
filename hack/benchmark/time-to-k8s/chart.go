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
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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
}

func main() {
	csvPath := flag.String("csv", "", "path to the CSV file")
	chartPath := flag.String("output", "", "path to output the chart to")
	flag.Parse()

	// map of the apps (minikube, kind, k3d) and their runs
	apps := make(map[string]runs)

	if err := readInCSV(*csvPath, apps); err != nil {
		log.Fatal(err)
	}

	values, totals, names := values(apps)

	if err := createChart(*chartPath, values, totals, names); err != nil {
		log.Fatal(err)
	}
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

		// 8-13 contain the run results
		for i := 8; i <= 13; i++ {
			v, err := strconv.ParseFloat(d[i], 64)
			if err != nil {
				return err
			}
			values = append(values, v)
		}
		newRun := run{values[0], values[1], values[2], values[3], values[4], values[5]}

		// get the app from the map and add the new run to it
		name := d[0]
		k, ok := apps[name]
		if !ok {
			k = runs{version: d[5]}
		}
		k.runs = append(k.runs, newRun)
		apps[name] = k
	}

	return nil
}

func values(apps map[string]runs) ([]plotter.Values, []float64, []string) {
	var cmdValues, apiValues, k8sValues, dnsSvcValues, appValues, dnsAnsValues plotter.Values
	names := []string{}
	totals := []float64{}

	// for each app, calculate the average for all the runs, and append them to the charting values
	for _, name := range []string{"minikube", "kind", "k3d"} {
		app := apps[name]
		var cmd, api, k8s, dnsSvc, appRun, dnsAns float64
		names = append(names, app.version)

		for _, l := range app.runs {
			cmd += l.cmd
			api += l.api
			k8s += l.k8s
			dnsSvc += l.dnsSvc
			appRun += l.app
			dnsAns += l.dnsAns
		}

		c := float64(len(app.runs))

		cmdAvg := cmd / c
		apiAvg := api / c
		k8sAvg := k8s / c
		dnsSvcAvg := dnsSvc / c
		appAvg := appRun / c
		dnsAnsAvg := dnsAns / c

		cmdValues = append(cmdValues, cmdAvg)
		apiValues = append(apiValues, apiAvg)
		k8sValues = append(k8sValues, k8sAvg)
		dnsSvcValues = append(dnsSvcValues, dnsSvcAvg)
		appValues = append(appValues, appAvg)
		dnsAnsValues = append(dnsAnsValues, dnsAnsAvg)

		total := cmdAvg + apiAvg + k8sAvg + dnsSvcAvg + appAvg + dnsAnsAvg
		totals = append(totals, total)
	}

	values := []plotter.Values{cmdValues, apiValues, k8sValues, dnsSvcValues, appValues, dnsAnsValues}

	return values, totals, names
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

	if err := p.Save(12*vg.Inch, 8*vg.Inch, chartPath); err != nil {
		return err
	}

	return nil
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
