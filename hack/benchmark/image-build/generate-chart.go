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
	"MinikubeDockerEnvContainerd",
	"MinikubeAddonRegistryContainerd",
	"MinikubeImageLoadCrio",
	"MinikubeImageCrio",
	"MinikubeAddonRegistryCrio",
	"Kind",
	"K3d",
	"Microk8s",
}

var RuntimeMethods = map[string][]string{
	"docker": {
		"MinikubeImageLoadDocker",
		"MinikubeImageBuild",
		"MinikubeDockerEnvDocker",
		"MinikubeAddonRegistryDocker",
	},

	"containerd": {
		"MinikubeImageLoadContainerd",
		"MinikubeImageContainerd",
		"MinikubeDockerEnvContainerd",
		"MinikubeAddonRegistryContainerd",
	},
}

const (
	INTERATIVE    = "Iterative"
	NONINTERATIVE = "NonIterative"
)

var Itrs = []string{
	INTERATIVE,
	// to simplify the output, non-interative is omitted
	// NONINTERATIVE,
}

// method name-> test result
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

type ItrTestResults struct {
	Date time.Time
	// itr name -> results
	Results map[string]ImageTestResults
}

type Records struct {
	Records []ItrTestResults
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
func readInLatestTestResult(latestBenchmarkPath string) ItrTestResults {

	var res = ItrTestResults{
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
		indicesInterative := []int{1, 5, 9, 13, 17, 21, 25, 29, 33, 37, 41, 45, 49, 53}
		// non-interative test results of each env are stored in the following columns
		indicesNonInterative := []int{3, 7, 11, 15, 19, 23, 27, 31, 35, 39, 43, 47, 51, 55}

		for _, i := range indicesInterative {
			if line[i] == "NaN" {
				// we use -1 as invalid value
				valuesInterative = append(valuesInterative, -1)
				continue
			}
			v, err := strconv.ParseFloat(line[i], 64)
			if err != nil {
				log.Fatal(err)
			}
			valuesInterative = append(valuesInterative, v)
		}

		for _, i := range indicesNonInterative {
			if line[i] == "NaN" {
				// we use -1 as invalid value
				valuesNonInterative = append(valuesNonInterative, -1)
				continue
			}
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

// updatePastTestResults overwrites the run file with the updated benchmarks list
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

	for _, itr := range Itrs {
		for _, image := range Images {
			createChart(record, itr, image, "docker", outputFolder)
			createChart(record, itr, image, "containerd", outputFolder)
		}
	}
}

func createChart(record Records, itr string, imageName string, runtime string, chartOutputPath string) {
	p := plot.New()
	p.Add(plotter.NewGrid())
	p.Legend.Top = true
	p.Title.Text = fmt.Sprintf("%s-%s-%s-performance", itr, imageName, runtime)
	p.X.Label.Text = "date"
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}
	p.Y.Label.Text = "time (seconds)"
	yMaxTotal := float64(0)

	// gonum plot do not have enough default colors in any group
	// so we combine different group of default colors
	colors := append([]color.Color{}, plotutil.SoftColors...)
	colors = append(colors, plotutil.DarkColors...)

	pointGroup := make(map[string]plotter.XYs)
	for _, name := range RuntimeMethods[runtime] {
		pointGroup[name] = make(plotter.XYs, 0)
	}

	for i := 0; i < len(record.Records); i++ {
		for _, method := range RuntimeMethods[runtime] {
			// for invalid values(<0) this point is dropped
			if v, ok := record.Records[i].Results[itr][imageName][method]; ok && v > 0 {
				point := plotter.XY{
					X: float64(record.Records[i].Date.Unix()),
					Y: record.Records[i].Results[itr][imageName][method],
				}
				pointGroup[method] = append(pointGroup[method], point)

				yMaxTotal = math.Max(yMaxTotal, point.Y)

			}
		}
	}
	p.Y.Max = yMaxTotal

	i := 0
	for method, xys := range pointGroup {
		line, points, err := plotter.NewLinePoints(xys)
		if err != nil {
			log.Fatal(err)
		}
		line.Color = colors[i]
		points.Color = colors[i]
		points.Shape = draw.CircleGlyph{}
		i++
		p.Add(line, points)
		p.Legend.Add(method, line)
	}

	filename := filepath.Join(chartOutputPath, fmt.Sprintf("%s_%s_%s_chart.png", itr, imageName, runtime))

	if err := p.Save(12*vg.Inch, 8*vg.Inch, filename); err != nil {
		log.Fatalf("failed creating png: %v", err)
	}
}
