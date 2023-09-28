/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var Images = []string{
	"buildpacksFewLargeFiles",
	// to simplify the output, the following images are omitted
	// "buildpacksFewSmallFiles",
	// "buildpacksManyLargeFiles",
	// "buildpacksManySmallFiles",
}

var Environments = []string{
	"MinikubeImageLoadDocker",
	"MinikubeImageBuild",
	"MinikubeDockerEnvDocker",
	"MinikubeAddonRegistryDocker",
	"MinikubeImageLoadContainerd",
	"MinikubeImageContainerd",
	"MinikubeAddonRegistryContainerd",
	"MinikubeImageLoadCrio",
	"MinikubeImageCrio",
	"MinikubeAddonRegistryCrio",
	"Kind",
	"K3d",
	"Microk8s",
}

var RuntimeEnvironments = map[string][]string{
	"docker": {
		"MinikubeImageLoadDocker",
		"MinikubeImageBuild",
		"MinikubeDockerEnvDocker",
		"MinikubeAddonRegistryDocker",
	},

	"containerd": {
		"MinikubeImageLoadContainerd",
		"MinikubeImageContainerd",
		"MinikubeAddonRegistryContainerd",
	},
}

const (
	INTERATIVE    = "Iterative"
	NONINTERATIVE = "NonIterative"
)

var Methods = []string{
	INTERATIVE,
	// to simplify the output, non-interative is omitted
	// NONINTERATIVE,
}

// env name-> test result
type TestResult map[string]float64

func NewTestResult(values []float64) TestResult {
	res := make(TestResult)
	for index, v := range values {
		res[Environments[index]] = v
	}
	return res
}

// imageName->TestResult
type ImageTestResults map[string]TestResult

type MethodTestResults struct {
	Date time.Time
	// method name -> results
	Results map[string]ImageTestResults
}

type Records struct {
	Records []MethodTestResults
}

func main() {
	latestTestResultPath := flag.String("csv", "", "path to the CSV file containing the latest benchmark result")
	pastTestRecordsPath := flag.String("past-runs", "", "path to the JSON file containing the past benchmark results")
	chartsPath := flag.String("charts", "", "path to the folder to write the daily charts to")
	flag.Parse()

	latestBenchmark := readInLatestTestResult(*latestTestResultPath)
	latestBenchmark.Date = time.Now()
	pastBenchmarks := readInPastTestResults(*pastTestRecordsPath)
	pastBenchmarks.Records = append(pastBenchmarks.Records, latestBenchmark)
	updatePastTestResults(pastBenchmarks, *pastTestRecordsPath)
	createDailyChart(pastBenchmarks, *chartsPath)
}

// readInLatestTestResult reads in the latest benchmark result from a CSV file
// and return the MethodTestResults object
func readInLatestTestResult(latestBenchmarkPath string) MethodTestResults {

	var res = MethodTestResults{
		Results: make(map[string]ImageTestResults),
	}
	res.Results[INTERATIVE] = make(ImageTestResults)
	res.Results[NONINTERATIVE] = make(ImageTestResults)

	f, err := os.Open(latestBenchmarkPath)
	if err != nil {
		log.Fatal(err)
	}

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
		if line[0] == "image" {
			continue
		}

		valuesInterative := []float64{}
		valuesNonInterative := []float64{}
		// interative test results of each env are stored in the following columns
		indicesInterative := []int{1, 5, 9, 13, 17, 21, 25, 29, 33, 37, 41, 45, 49}
		// non-interative test results of each env are stored in the following columns
		indicesNonInterative := []int{3, 7, 11, 15, 19, 23, 27, 31, 35, 39, 43, 47, 51}

		for _, i := range indicesInterative {
			v, err := strconv.ParseFloat(line[i], 64)
			if err != nil {
				log.Fatal(err)
			}
			valuesInterative = append(valuesInterative, v)
		}

		for _, i := range indicesNonInterative {
			v, err := strconv.ParseFloat(line[i], 64)
			if err != nil {
				log.Fatal(err)
			}
			valuesNonInterative = append(valuesNonInterative, v)
		}

		imageName := line[0]

		res.Results[INTERATIVE][imageName] = NewTestResult(valuesInterative)
		res.Results[NONINTERATIVE][imageName] = NewTestResult(valuesNonInterative)

	}

	return res
}

// readInPastTestResults reads in the past benchmark results from a JSON file
func readInPastTestResults(pastTestRecordPath string) Records {

	record := Records{}
	data, err := os.ReadFile(pastTestRecordPath)
	if os.IsNotExist(err) {
		return record
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(data, &record); err != nil {
		log.Fatal(err)
	}

	return record
}

// updateRunsFile overwrites the run file with the updated benchmarks list
func updatePastTestResults(h Records, pastTestRecordPath string) {
	b, err := json.Marshal(h)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(pastTestRecordPath, b, 0600); err != nil {
		log.Fatal(err)
	}
}
func createDailyChart(record Records, outputFolder string) {

	for _, method := range Methods {
		for _, image := range Images {
			createChart(record, method, image, "docker", outputFolder)
			createChart(record, method, image, "containerd", outputFolder)
		}
	}
}

func createChart(record Records, methodName string, imageName string, runtime string, chartOutputPath string) {
	p := plot.New()
	p.Add(plotter.NewGrid())
	p.Legend.Top = true
	p.Title.Text = fmt.Sprintf("%s-%s-%s-performance", methodName, imageName, runtime)
	p.X.Label.Text = "date"
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}
	p.Y.Label.Text = "time (seconds)"
	yMaxTotal := float64(0)

	// gonum plot do not have enough default colors in any group
	// so we combine different group of default colors
	colors := append([]color.Color{}, plotutil.SoftColors...)
	colors = append(colors, plotutil.DarkColors...)

	pointGroup := make(map[string]plotter.XYs)
	for _, name := range RuntimeEnvironments[runtime] {
		pointGroup[name] = make(plotter.XYs, len(record.Records))

	}

	for i := 0; i < len(record.Records); i++ {
		for _, envName := range RuntimeEnvironments[runtime] {
			pointGroup[envName][i].X = float64(record.Records[i].Date.Unix())
			pointGroup[envName][i].Y = record.Records[i].Results[methodName][imageName][envName]
			yMaxTotal = math.Max(yMaxTotal, pointGroup[envName][i].Y)
		}
	}
	p.Y.Max = yMaxTotal

	i := 0
	for envName, xys := range pointGroup {
		line, points, err := plotter.NewLinePoints(xys)
		if err != nil {
			log.Fatal(err)
		}
		line.Color = colors[i]
		points.Color = colors[i]
		points.Shape = draw.CircleGlyph{}
		i++
		p.Add(line, points)
		p.Legend.Add(envName, line)
	}

	filename := filepath.Join(chartOutputPath, fmt.Sprintf("%s_%s_%s_chart.png", methodName, imageName, runtime))

	if err := p.Save(12*vg.Inch, 8*vg.Inch, filename); err != nil {
		log.Fatalf("failed creating png: %v", err)
	}
}
