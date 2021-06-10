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
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	dataCsv   = flag.String("data-csv", "", "Source data to compute flake rates on")
	dateRange = flag.Uint("date-range", 5, "Number of test dates to consider when computing flake rate")
)

func main() {
	flag.Parse()

	file, err := os.Open(*dataCsv)
	if err != nil {
		exit("Unable to read data CSV", err)
	}

	testEntries := readData(file)
	splitEntries := splitData(testEntries)
	filteredEntries := filterRecentEntries(splitEntries, *dateRange)
	flakeRates := computeFlakeRates(filteredEntries)
	averageDurations := computeAverageDurations(filteredEntries)
	fmt.Println("Environment,Test,Flake Rate,Duration")
	for environment, environmentSplit := range flakeRates {
		for test, flakeRate := range environmentSplit {
			duration := averageDurations[environment][test]
			fmt.Printf("%s,%s,%.2f,%.3f\n", environment, test, flakeRate*100, duration)
		}
	}
}

// One entry of a test run.
// Example: TestEntry {
//	 name: "TestFunctional/parallel/LogsCmd",
//   environment: "Docker_Linux",
//	 date: time.Now,
//   status: "Passed",
//   duration: 0.1,
// }
type testEntry struct {
	name        string
	environment string
	date        time.Time
	status      string
	duration    float32
}

// A map with keys of (environment, test_name) to values of slcies of TestEntry.
type splitEntryMap map[string]map[string][]testEntry

// Reads CSV `file` and consumes each line to be a single TestEntry.
func readData(file io.Reader) []testEntry {
	testEntries := []testEntry{}

	fileReader := bufio.NewReaderSize(file, 256)
	previousLine := []string{"", "", "", "", "", ""}
	firstLine := true
	for {
		lineBytes, _, err := fileReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			exit("Error reading data CSV", err)
		}
		line := string(lineBytes)
		fields := strings.Split(line, ",")
		if firstLine {
			if len(fields) != 6 {
				exit(fmt.Sprintf("Data CSV in incorrect format. Expected 6 columns, but got %d", len(fields)), fmt.Errorf("bad CSV format"))
			}
			firstLine = false
		}
		for i, field := range fields {
			if field == "" {
				fields[i] = previousLine[i]
			}
		}
		if len(fields) != 6 {
			fmt.Printf("Found line with wrong number of columns. Expectd 6, but got %d - skipping\n", len(fields))
			continue
		}
		previousLine = fields
		if fields[4] == "Passed" || fields[4] == "Failed" {
			date, err := time.Parse("2006-01-02", fields[1])
			if err != nil {
				fmt.Printf("Failed to parse date: %v\n", err)
				continue
			}
			duration, err := strconv.ParseFloat(fields[5], 32)
			if err != nil {
				fmt.Printf("Failed to parse duration: %v\n", err)
				continue
			}
			testEntries = append(testEntries, testEntry{
				name:        fields[3],
				environment: fields[2],
				date:        date,
				status:      fields[4],
				duration:    float32(duration),
			})
		}
	}
	return testEntries
}

// Splits `testEntries` up into maps indexed first by environment and then by test.
func splitData(testEntries []testEntry) splitEntryMap {
	splitEntries := make(splitEntryMap)

	for _, entry := range testEntries {
		appendEntry(splitEntries, entry.environment, entry.name, entry)
	}

	return splitEntries
}

// Appends `entry` to `splitEntries` at the `environment` and `test`.
func appendEntry(splitEntries splitEntryMap, environment, test string, entry testEntry) {
	// Lookup the environment.
	environmentSplit, ok := splitEntries[environment]
	if !ok {
		// If the environment map is missing, make a map for this environment and store it.
		environmentSplit = make(map[string][]testEntry)
		splitEntries[environment] = environmentSplit
	}

	// Lookup the test.
	testSplit, ok := environmentSplit[test]
	if !ok {
		// If the test is missing, make a slice for this test.
		testSplit = make([]testEntry, 0)
		// The slice is not inserted, since it will be replaced anyway.
	}
	environmentSplit[test] = append(testSplit, entry)
}

// Filters `splitEntries` to include only the most recent `date_range` dates.
func filterRecentEntries(splitEntries splitEntryMap, dateRange uint) splitEntryMap {
	filteredEntries := make(splitEntryMap)

	for environment, environmentSplit := range splitEntries {
		for test, testSplit := range environmentSplit {
			dates := make([]time.Time, len(testSplit))
			for _, entry := range testSplit {
				dates = append(dates, entry.date)
			}
			// Sort dates from future to past.
			sort.Slice(dates, func(i, j int) bool {
				return dates[j].Before(dates[i])
			})
			datesInRange := make([]time.Time, 0, dateRange)
			var lastDate time.Time
			// Go through each date.
			for _, date := range dates {
				// If date is the same as last date, ignore it.
				if date.Equal(lastDate) {
					continue
				}

				// Add the date.
				datesInRange = append(datesInRange, date)
				lastDate = date
				// If the date_range has been hit, break out.
				if uint(len(datesInRange)) == dateRange {
					break
				}
			}

			for _, entry := range testSplit {
				// Look for the first element <= entry.date
				index := sort.Search(len(datesInRange), func(i int) bool {
					return !datesInRange[i].After(entry.date)
				})
				// If no date is <= entry.date, or the found date does not equal entry.date.
				if index == len(datesInRange) || !datesInRange[index].Equal(entry.date) {
					continue
				}
				appendEntry(filteredEntries, environment, test, entry)
			}
		}
	}
	return filteredEntries
}

// Computes the flake rates over each entry in `splitEntries`.
func computeFlakeRates(splitEntries splitEntryMap) map[string]map[string]float32 {
	flakeRates := make(map[string]map[string]float32)
	for environment, environmentSplit := range splitEntries {
		for test, testSplit := range environmentSplit {
			failures := 0
			for _, entry := range testSplit {
				if entry.status == "Failed" {
					failures++
				}
			}
			setValue(flakeRates, environment, test, float32(failures)/float32(len(testSplit)))
		}
	}
	return flakeRates
}

// Computes the average durations over each entry in `splitEntries`.
func computeAverageDurations(splitEntries splitEntryMap) map[string]map[string]float32 {
	averageDurations := make(map[string]map[string]float32)
	for environment, environmentSplit := range splitEntries {
		for test, testSplit := range environmentSplit {
			durationSum := float32(0)
			for _, entry := range testSplit {
				durationSum += entry.duration
			}
			if len(testSplit) != 0 {
				durationSum /= float32(len(testSplit))
			}
			setValue(averageDurations, environment, test, durationSum)
		}
	}
	return averageDurations
}

// Sets the `value` of keys `environment` and `test` in `mapEntries`.
func setValue(mapEntries map[string]map[string]float32, environment, test string, value float32) {
	// Lookup the environment.
	environmentRates, ok := mapEntries[environment]
	if !ok {
		// If the environment map is missing, make a map for this environment and store it.
		environmentRates = make(map[string]float32)
		mapEntries[environment] = environmentRates
	}
	environmentRates[test] = value
}

// exit will exit and clean up minikube
func exit(msg string, err error) {
	fmt.Printf("WithError(%s)=%v called from:\n%s", msg, err, debug.Stack())
	os.Exit(60)
}
