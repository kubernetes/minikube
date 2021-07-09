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
	"image/color"
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

var FOLDER = "site/static/images/benchmarks/cpuUsage/autoPause"

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
		p.Title.Text = "CPU% Busy Overhead - With Auto Pause vs. Non Auto Pause (less is better)"
	} else if runtime.GOOS == "linux" {
		p.Title.Text = "CPU% Busy Overhead - With Auto Pause vs. Non Auto Pause (less is better)"
	}
	p.Y.Label.Text = "CPU overhead%"

	// Open non-autopause csv file of benchmark summary
	napResults := []float64{}
	var napFn string = "./out/benchmark-results/" + sessionID + "/cstat.nonautopause.summary"
	napFile, err := os.Open(napFn)
	if err != nil {
		return errors.Wrap(err, "Missing summary csv")
	}
	defer napFile.Close()

	// Read result values from benchmark summary csv
	napReader := csv.NewReader(napFile)
	var napLine []string
	for {
		napLine, err = napReader.Read()
		if err != nil {
			break
		}

		s, err := strconv.ParseFloat(napLine[0], 64)
		if err != nil {
			return errors.Wrap(err, "Failed to convert to float64")
		}
		napResults = append(napResults, s)
	}

	// Open auto-pause csv file of benchmark summary
	apResults := []float64{}
	var apFn string = "./out/benchmark-results/" + sessionID + "/cstat.autopause.summary"
	apFile, err := os.Open(apFn)
	if err != nil {
		return errors.Wrap(err, "Missing summary csv")
	}
	defer apFile.Close()

	// Read result values from benchmark summary csv
	apReader := csv.NewReader(apFile)
	var apLine []string
	for {
		apLine, err = apReader.Read()
		if err != nil {
			break
		}

		s, err := strconv.ParseFloat(apLine[0], 64)
		if err != nil {
			return errors.Wrap(err, "Failed to convert to float64")
		}
		apResults = append(apResults, s)
	}

	// Set bar graph width
	breadth := vg.Points(40)

	// Create Bar instance with non-autopause benchmark results
	barNAP, err := plotter.NewBarChart(plotter.Values(napResults), breadth)
	if err != nil {
		return errors.Wrap(err, "Failed to create bar chart")
	}

	// Set border of the bar graph. 0 is no border color
	barNAP.LineStyle.Width = vg.Length(0)
	// Add bar name
	p.Legend.Add("Initial Start CPU usage Before Pause", barNAP)
	// Set bar color to gray.
	barNAP.Color = color.RGBA{184, 184, 184, 255}

	// Create Bar instance with auto-pause benchmark results
	barAP, err := plotter.NewBarChart(plotter.Values(apResults), breadth)
	if err != nil {
		return errors.Wrap(err, "Failed to create bar chart")
	}

	// Set border of the bar graph. 0 is no border color
	barAP.LineStyle.Width = vg.Length(0)
	// Add bar name
	p.Legend.Add("Auto Paused CPU usage", barAP)
	// Set bar color. 1 is green
	barAP.Color = plotutil.Color(1)

	hb := vg.Points(20)
	barNAP.Offset = -hb
	barAP.Offset = hb
	p.Add(barNAP, barAP)

	// Set legend position upper
	p.Legend.Top = true

	// Add x-lay names
	if runtime.GOOS == "darwin" {
		p.NominalX("OS idle", "minikube hyperkit", "minikube virtualbox", "minikube docker", "Docker for Mac Kubernetes", "k3d", "kind")
	} else if runtime.GOOS == "linux" {
		p.NominalX("OS idle", "minikube kvm2", "minikube virtualbox", "minikube docker", "Docker idle", "k3d", "kind")
	}

	// Set non-autopause data label to each bar
	var napLabels []string
	for i := range napResults {
		nLabel := strconv.FormatFloat(napResults[i], 'f', -1, 64)
		napLabels = append(napLabels, nLabel)
	}

	var napCPU []plotter.XY
	for i := range napResults {
		napXPos := float64(i) - 0.25
		napYPos := napResults[i] + 0.1
		napXY := plotter.XY{X: napXPos, Y: napYPos}
		napCPU = append(napCPU, napXY)
	}
	// CPU Busy% non-autopause data label
	napl, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    napCPU,
		Labels: napLabels,
	},
	)
	if err != nil {
		return err
	}

	// Set auto-pause data label to each bar
	var apLabels []string
	for i := range apResults {
		if apResults[i] == 0 {
			apLabels = append(apLabels, "N/A")
		} else {
			apLabel := strconv.FormatFloat(apResults[i], 'f', -1, 64)
			apLabels = append(apLabels, apLabel)
		}
	}

	var apCPU []plotter.XY
	for i := range apResults {
		apXPos := float64(i) + 0.05
		apYPos := apResults[i] + 0.1
		apXY := plotter.XY{X: apXPos, Y: apYPos}
		apCPU = append(apCPU, apXY)
	}
	// CPU Busy% auto-pause data label
	apl, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    apCPU,
		Labels: apLabels,
	},
	)
	if err != nil {
		return err
	}

	// define max cpu busy% to 20%
	p.Y.Max = 20
	p.Y.Tick.Marker = integerTicks{}
	// Add CPU Busy% label to plot
	p.Add(napl, apl)

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
