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
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var FOLDER = "site/static/images/benchmarks/cpuUsage/idleOnly"

type integerTicks struct{}

func (integerTicks) Ticks(min, max float64) []plot.Tick {
	var t []plot.Tick
	for i := math.Trunc(min); i <= max; i += 50 {
		t = append(t, plot.Tick{Value: i, Label: fmt.Sprint(i)})
	}
	return t
}

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute() error {
	// sessionID is generated and used at cpu usage benchmark
	sessionID := os.Args[1]
	if len(sessionID) == 0 {
		return errors.New("Please identify sessionID")
	}

	// Create plot instance
	p := plot.New()

	// Set view options
	if runtime.GOOS == "darwin" {
		p.Title.Text = "CPU% Busy Overhead - Average first 5 minutes on macOS (less is better)"
	} else if runtime.GOOS == "linux" {
		p.Title.Text = "CPU% Busy Overhead - Average first 5 minutes on Linux (less is better)"
	}
	p.Y.Label.Text = "CPU overhead%"

	// Open csv file of benchmark summary
	results := []float64{}
	var fn string = "./out/benchmark-results/" + sessionID + "/cstat.summary"
	file, err := os.Open(fn)
	if err != nil {
		return errors.Wrap(err, "Missing summary csv")
	}
	defer file.Close()

	// Read result values from benchmark summary csv
	reader := csv.NewReader(file)
	var line []string
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}

		s, err := strconv.ParseFloat(line[0], 64)
		if err != nil {
			return errors.Wrap(err, "Failed to convert to float64")
		}
		results = append(results, s)
	}

	// Set bar graph width
	breadth := vg.Points(40)
	// Create Bar instance with benchmark results
	bar, err := plotter.NewBarChart(plotter.Values(results), breadth)
	if err != nil {
		return errors.Wrap(err, "Failed to create bar chart")
	}

	// Set border of the bar graph. 0 is no border color
	bar.LineStyle.Width = vg.Length(0)
	// Add bar name
	p.Legend.Add("CPU Busy%", bar)
	// Set bar color. 2 is blue
	bar.Color = plotutil.Color(2)
	p.Add(bar)

	// Set legend position upper
	p.Legend.Top = true

	// Add x-lay names
	if runtime.GOOS == "darwin" {
		p.NominalX("OS idle", "minikube hyperkit", "minikube virtualbox", "minikube docker", "Docker for Mac Kubernetes", "k3d", "kind")
	} else if runtime.GOOS == "linux" {
		p.NominalX("OS idle", "minikube kvm2", "minikube virtualbox", "minikube docker", "Docker idle", "k3d", "kind")
	}

	// Set data label to each bar
	var cpuLabels []string
	for i := range results {
		cLabel := strconv.FormatFloat(results[i], 'f', -1, 64)
		cpuLabels = append(cpuLabels, cLabel)
	}

	var xysCPU []plotter.XY
	for i := range results {
		rxPos := float64(i) - 0.13
		ryPos := results[i] + 0.1
		cXY := plotter.XY{X: rxPos, Y: ryPos}
		xysCPU = append(xysCPU, cXY)
	}
	// CPU Busy% data label
	cl, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    xysCPU,
		Labels: cpuLabels,
	},
	)
	if err != nil {
		return err
	}

	// define max cpu busy% to 20%
	p.Y.Max = 20
	p.Y.Tick.Marker = integerTicks{}
	// Add CPU Busy% label to plot
	p.Add(cl)

	// Output bar graph
	if runtime.GOOS == "darwin" {
		if err := p.Save(13*vg.Inch, 8*vg.Inch, FOLDER+"/mac.png"); err != nil {
			return errors.Wrap(err, "Failed to create bar graph png")
		}
		log.Printf("Generated graph png to %s/mac.png", FOLDER)
	} else if runtime.GOOS == "linux" {
		if err := p.Save(13*vg.Inch, 10*vg.Inch, FOLDER+"/linux.png"); err != nil {
			return errors.Wrap(err, "Failed to create bar graph png")
		}
		log.Printf("Generated graph png to %s/linux.png", FOLDER)
	}
	return nil
}
