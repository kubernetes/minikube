/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
)

type ShortSummary struct {
	NumberOfTests int
	NumberOfFail  int
	NumberOfPass  int
	NumberOfSkip  int
	FailedTests   []string
	PassedTests   []string
	SkippedTests  []string
	Durations     map[string]float64
	TotalDuration float64
	GopoghVersion string
	GopoghBuild   string
	Detail        struct {
		Name     string
		Details  string
		PR       string
		RepoName string
	}
}

// ParseEnvironmentList read the existing environments from the file
func ParseEnvironmentList(listFile string) ([]string, error) {
	data, err := os.ReadFile(listFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n"), nil
}

func TestSummariesFromGCP(pr, rootJob string, envList []string, client *storage.Client) (map[string]*ShortSummary, error) {
	envToSummaries := map[string]*ShortSummary{}
	for _, env := range envList {
		if summary, err := getTestSummaryFromGCP(pr, rootJob, env, client); err == nil {
			if summary != nil {
				// if the summary is nil(missing) we just skip it
				envToSummaries[env] = summary
			}
		} else {
			return nil, fmt.Errorf("failed to fetch %s test summary from gcp, err: %v", env, err)
		}
	}
	return envToSummaries, nil
}

// getFromSummary get the summary of a test on the specified env from the specified summary.
func getTestSummaryFromGCP(pr, rootJob, env string, client *storage.Client) (*ShortSummary, error) {
	ctx := context.TODO()

	btk := client.Bucket("minikube-builds")
	obj := btk.Object(fmt.Sprintf("logs/%s/%s/%s_summary.json", pr, rootJob, env))

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			// if this file does not exist, just skip it
			return nil, nil
		}
		return nil, err
	}
	// read the file
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var summary ShortSummary
	if err = json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to deserialize the file: %v", err)
	}
	return &summary, nil

}

// GetFlakeRate downloaded recent flake rate from gcs, and return the map{env->map{testname->flake rate}}
func GetFlakeRate(client *storage.Client) (map[string]map[string]float64, error) {
	btk := client.Bucket("minikube-flake-rate")
	obj := btk.Object("flake_rates.csv")
	reader, err := obj.NewReader(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to read the flake rate file: %v", err)
	}
	// parse the csv file to the map
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read and parse the flake rate file: %v", err)
	}
	result := map[string]map[string]float64{}
	for i := 1; i < len(records); i++ {
		// for each line in csv we extract env, test name and flake rate
		env := records[i][0]
		test := records[i][1]
		flakeRate, err := strconv.ParseFloat(records[i][2], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the flake rate file at line %d: %v", i+1, err)
		}
		if _, ok := result[env]; !ok {
			result[env] = make(map[string]float64, 0)
		}
		result[env][test] = flakeRate
	}
	return result, nil
}

func GenerateCommentMessage(summaries map[string]*ShortSummary, flakeRates map[string]map[string]float64, pr, rootJob string) string {
	//builder := strings.Builder{}
	type failedTest struct {
		flakeRate float64
		env       string
		testName  string
	}
	// for each environment, we sort failed tests according to the flake rate of that test on master branch
	envFailedTestList := map[string][]failedTest{}
	for env, summary := range summaries {
		failedTestList := []failedTest{}
		for _, test := range summary.FailedTests {
			// if we cannot find the test, we assign the flake rate as -1, meaning N/A
			flakerate := -1.0
			if _, ok := flakeRates[env]; ok {
				if v, ok := flakeRates[env][test]; ok {
					flakerate = v
				}
			}
			failedTestList = append(failedTestList,
				failedTest{
					flakeRate: flakerate,
					env:       env,
					testName:  test,
				})
		}

		sort.Slice(failedTestList, func(i, j int) bool {
			return failedTestList[i].flakeRate < failedTestList[j].flakeRate
		})
		envFailedTestList[env] = failedTestList
	}
	// we convert the result into a 2d string slice representing a markdown table,
	// whose each line represents a line of the table
	table := [][]string{
		// title of the table
		{"Environment", "Test Name", "Flake Rate"},
	}
	// if an env has too much failures we will just skip it and print a message in the end
	tooMuchFailure := []string{}
	for env, list := range envFailedTestList {
		if len(list) > MAX_ITEM_ENV {
			tooMuchFailure = append(tooMuchFailure, env)
			continue
		}
		for _, item := range list {
			flakeRateString := fmt.Sprintf("%.2f%% %s", item.flakeRate, testFlakeChartMDLink(env, item.testName))
			if item.flakeRate < 0 {
				flakeRateString = "Unknown"
			}
			table = append(table, []string{
				envChartMDLink(env, len(list)),
				item.testName + gopoghMDLink(pr, rootJob, env, item.testName),
				flakeRateString,
			})
		}
	}

	builder := strings.Builder{}
	builder.WriteString(
		fmt.Sprintf("Here are the number of top %d failed tests in each environments with lowest flake rate.\n\n", MAX_ITEM_ENV))
	builder.WriteString(generateMarkdownTable(table))
	if len(tooMuchFailure) > 0 {

		builder.WriteString("\n\n Besides the following environments have too much failed tests:")
		for _, env := range tooMuchFailure {
			builder.WriteString(fmt.Sprintf("\n\n - %s: %d failed", env, len(envFailedTestList[env])))
		}
	}
	builder.WriteString("\n\nTo see the flake rates of all tests by environment, click [here](https://minikube.sigs.k8s.io/docs/contrib/test_flakes/).")
	return builder.String()
}
func envChartMDLink(env string, failedTestNumber int) string {
	return fmt.Sprintf("[%s (%d failed)](https://gopogh-server-tts3vkcpgq-uc.a.run.app/?env=%s)", env, failedTestNumber, env)
}

func testFlakeChartMDLink(env string, testName string) string {
	return fmt.Sprintf("[(chart)](https://gopogh-server-tts3vkcpgq-uc.a.run.app/?env=%s&test=%s)", env, testName)
}

func gopoghMDLink(pr, rootJob, env, testName string) string {
	return fmt.Sprintf("[(gopogh)](https://storage.googleapis.com/minikube-builds/logs/%s/%s/%s.html#%s)", pr, rootJob, env, testName)
}

// generateMarkdownTable convert 2d string slice into markdown table. The first string slice is the header of the table
func generateMarkdownTable(table [][]string) string {
	builder := strings.Builder{}
	for i, group := range table {
		builder.WriteString("|")
		for j := 0; j < len(group); j++ {
			builder.WriteString(group[j])
			builder.WriteString("|")
		}
		builder.WriteString("\n")

		if i != 0 {
			continue
		}
		// generate the hyphens seperator
		builder.WriteString("|")
		for j := 0; j < len(group); j++ {
			builder.WriteString(" ---- |")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n\n")
	return builder.String()

}
