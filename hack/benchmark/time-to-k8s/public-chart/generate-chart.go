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
	"encoding/json"
	"flag"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// benchmark contains the duration of the benchmark steps
type benchmark struct {
	Date   time.Time `json:"date"`
	Cmd    float64   `json:"cmd"`
	API    float64   `json:"api"`
	K8s    float64   `json:"k8s"`
	DNSSvc float64   `json:"dnsSvc"`
	App    float64   `json:"app"`
	DNSAns float64   `json:"dnsAns"`
	Total  float64   `json:"total"`
}

// benchmarks contains a list of benchmarks, used for storing benchmark results to JSON
type benchmarks struct {
	Benchmarks []benchmark `json:"benchmarks"`
}

func main() {
	latestBenchmarkPath := flag.String("csv", "", "path to the CSV file containing the latest benchmark result")
	chartOutputPath := flag.String("output", "", "path to output the chart to")
	pastBenchmarksPath := flag.String("past-runs", "", "path to the JSON file containing the past benchmark results")
	flag.Parse()

	latestBenchmark := readInLatestBenchmark(*latestBenchmarkPath)
	pastBenchmarks := readInPastBenchmarks(*pastBenchmarksPath)
	pastBenchmarks.Benchmarks = append(pastBenchmarks.Benchmarks, latestBenchmark)
	updateRunsFile(pastBenchmarks, *pastBenchmarksPath)
	createChart(pastBenchmarks.Benchmarks, *chartOutputPath)
}

// readInLatestBenchmark reads in the latest benchmark result from a CSV file
func readInLatestBenchmark(latestBenchmarkPath string) benchmark {
	f, err := os.Open(latestBenchmarkPath)
	if err != nil {
		log.Fatal(err)
	}

	var cmd, api, k8s, dnsSvc, app, dnsAns float64
	steps := []*float64{&cmd, &api, &k8s, &dnsSvc, &app, &dnsAns}
	count := 0

	r := csv.NewReader(f)
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// skip the first line of the CSV file
		if line[0] == "name" {
			continue
		}

		values := []float64{}

		// 8-13 contain the benchmark results
		for i := 8; i <= 13; i++ {
			v, err := strconv.ParseFloat(line[i], 64)
			if err != nil {
				log.Fatal(err)
			}
			values = append(values, v)
		}
		count++
		for i, step := range steps {
			*step += values[i]
		}
	}

	var total float64
	for _, step := range steps {
		*step /= float64(count)
		total += *step
	}

	return benchmark{time.Now(), cmd, api, k8s, dnsSvc, app, dnsAns, total}
}

// readInPastBenchmarks reads in the past benchmark results from a JSON file
func readInPastBenchmarks(pastBenchmarksPath string) *benchmarks {
	data, err := os.ReadFile(pastBenchmarksPath)
	if err != nil {
		log.Fatal(err)
	}

	b := &benchmarks{}
	if err := json.Unmarshal(data, b); err != nil {
		log.Fatal(err)
	}

	return b
}

// updateRunsFile overwrites the run file with the updated benchmarks list
func updateRunsFile(h *benchmarks, pastRunsPath string) {
	b, err := json.Marshal(h)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(pastRunsPath, b, 0600); err != nil {
		log.Fatal(err)
	}
}

// createChart creates a time series chart of the benchmarks
func createChart(benchmarks []benchmark, chartOutputPath string) {
	n := len(benchmarks)
	var cmdXYs, apiXYs, k8sXYs, dnsSvcXYs, appXYs, dnsAnsXYs, totalXYs plotter.XYs
	xys := []*plotter.XYs{&cmdXYs, &apiXYs, &k8sXYs, &dnsSvcXYs, &appXYs, &dnsAnsXYs, &totalXYs}

	for _, xy := range xys {
		*xy = make(plotter.XYs, n)
	}

	for i, b := range benchmarks {
		date := float64(b.Date.Unix())
		xyValues := []struct {
			xys   *plotter.XYs
			value float64
		}{
			{&cmdXYs, b.Cmd},
			{&apiXYs, b.API},
			{&k8sXYs, b.K8s},
			{&dnsSvcXYs, b.DNSSvc},
			{&appXYs, b.App},
			{&dnsAnsXYs, b.DNSAns},
			{&totalXYs, b.Total},
		}
		for _, xyValue := range xyValues {
			xy := &(*xyValue.xys)[i]
			xy.Y = xyValue.value
			xy.X = date
		}
	}

	p := plot.New()
	p.Add(plotter.NewGrid())
	p.Legend.Top = true
	p.Title.Text = "time-to-k8s"
	p.X.Label.Text = "date"
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}
	p.Y.Label.Text = "time (seconds)"
	p.Y.Max = 95

	steps := []struct {
		xys   plotter.XYs
		rgba  color.RGBA
		label string
	}{
		{cmdXYs, color.RGBA{R: 255, A: 255}, "Command Exec"},
		{apiXYs, color.RGBA{G: 255, A: 255}, "API Server Answering"},
		{k8sXYs, color.RGBA{B: 255, A: 255}, "Kubernetes SVC"},
		{dnsSvcXYs, color.RGBA{R: 255, B: 255, A: 255}, "DNS SVC"},
		{appXYs, color.RGBA{R: 255, G: 255, A: 255}, "App Running"},
		{dnsAnsXYs, color.RGBA{G: 255, B: 255, A: 255}, "DNS Answering"},
		{totalXYs, color.RGBA{B: 255, R: 140, A: 255}, "Total"},
	}

	for _, step := range steps {
		line, points := newLinePoints(step.xys, step.rgba)
		p.Add(line, points)
		p.Legend.Add(step.label, line)
	}

	if err := p.Save(12*vg.Inch, 8*vg.Inch, chartOutputPath); err != nil {
		log.Fatal(err)
	}
}

func newLinePoints(xys plotter.XYs, lineColor color.RGBA) (*plotter.Line, *plotter.Scatter) {
	line, points, err := plotter.NewLinePoints(xys)
	if err != nil {
		log.Fatal(err)
	}
	line.Color = lineColor
	points.Color = lineColor
	points.Shape = draw.CircleGlyph{}

	return line, points
}
