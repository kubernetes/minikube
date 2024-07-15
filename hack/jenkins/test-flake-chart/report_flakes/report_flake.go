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
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
)

const unknown float64 = -1.0

type shortSummary struct {
	NumberOfTests int                `json:"NumberOfTests"`
	NumberOfFail  int                `json:"NumberOfFail"`
	NumberOfPass  int                `json:"NumberOfPass"`
	NumberOfSkip  int                `json:"NumberOfSkip"`
	FailedTests   []string           `json:"FailedTests"`
	PassedTests   []string           `json:"PassedTests"`
	SkippedTests  []string           `json:"SkippedTests"`
	Durations     map[string]float64 `json:"Durations"`
	TotalDuration float64            `json:"TotalDuration"`
	GopoghVersion string             `json:"GopoghVersion"`
	GopoghBuild   string             `json:"GopoghBuild"`
	Detail        struct {
		Name     string `json:"Name"`
		Details  string `json:"Details"`
		PR       string `json:"PR"`
		RepoName string `json:"RepoName"`
	} `json:"Detail"`
}

// parseEnvironmentList reads the existing environments from the file
func parseEnvironmentList(listFile string) ([]string, error) {
	data, err := os.ReadFile(listFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n"), nil
}

func testSummariesFromGCP(ctx context.Context, pr, rootJob string, envList []string, client *storage.Client) (map[string]*shortSummary, error) {
	envToSummaries := map[string]*shortSummary{}
	for _, env := range envList {
		summary, err := testSummaryFromGCP(ctx, pr, rootJob, env, client)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s test summary from gcp, err: %v", env, err)
		}
		if summary != nil {
			// if the summary is nil(missing) we just skip it
			envToSummaries[env] = summary
		}
	}
	return envToSummaries, nil
}

// testSummaryFromGCP gets the summary of a test for the specified env.
func testSummaryFromGCP(ctx context.Context, pr, rootJob, env string, client *storage.Client) (*shortSummary, error) {

	btk := client.Bucket("minikube-builds")
	obj := btk.Object(fmt.Sprintf("logs/%s/%s/%s_summary.json", pr, rootJob, env))

	reader, err := obj.NewReader(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		// if this file does not exist, just skip it
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// read the file
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var summary shortSummary
	if err = json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to deserialize the file: %v", err)
	}
	return &summary, nil

}

// flakeRate downloads recent flake rates from GCS, and returns a map{env->map{testname->flake rate}}
func flakeRate(ctx context.Context, client *storage.Client) (map[string]map[string]float64, error) {
	btk := client.Bucket("minikube-flake-rate")
	obj := btk.Object("flake_rates.csv")
	reader, err := obj.NewReader(ctx)
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
		if len(records[i]) < 3 {
			// the csv must have at least 2 columns
			continue
		}
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

func generateCommentMessage(summaries map[string]*shortSummary, flakeRates map[string]map[string]float64, pr, rootJob string) string {
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
			// if we cannot find the test, we assign the flake rate
			// as -1, meaning N/A
			flakerate := unknown
			if v, ok := flakeRates[env][test]; ok {
				flakerate = v
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
	// we convert the result into a 2d string slice representing a markdown
	// table, whose each line represents a line of the table
	table := [][]string{
		// title of the table
		{"Environment", "Test Name", "Flake Rate"},
	}
	// if an env has too many failures we will just skip it and print a
	// message in the end if the failed tests have high flake rates (over
	// 50% for all), it will also be skipped in the table
	foldedFailures := []string{}
	for env, list := range envFailedTestList {
		if len(list) > maxItemEnv {
			foldedFailures = append(foldedFailures, env)
			continue
		}
		for i, item := range list {
			if item.flakeRate > 50 {
				if i == 0 {
					// if this is the first failed test that
					// means each tests in this env has a
					// flakerate>50% and all of them will
					// not be shown in the table
					foldedFailures = append(foldedFailures, env)
				}
				break
			}
			flakeRateString := fmt.Sprintf("%.2f%% %s", item.flakeRate, testFlakeChartMDLink(env, item.testName))
			if item.flakeRate == unknown {
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
		fmt.Sprintf("Here are the number of top %d failed tests in each environments with lowest flake rate.\n\n", maxItemEnv))
	builder.WriteString(generateMarkdownTable(table))
	if len(foldedFailures) > 0 {

		builder.WriteString("\n\n Besides the following environments also have failed tests:")
		for _, env := range foldedFailures {
			builder.WriteString(fmt.Sprintf("\n\n - %s: %d failed %s ", env, len(envFailedTestList[env]), gopoghMDLink(pr, rootJob, env, "")))
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

// generateMarkdownTable convert 2d string slice into markdown table. The first
// string slice is the header of the table
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
		// generate the hyphens separator
		builder.WriteString("|")
		for j := 0; j < len(group); j++ {
			builder.WriteString(" ---- |")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n\n")
	return builder.String()

}
