/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

	testEntries := ReadData(file)
	for _, entry := range testEntries {
		fmt.Printf("Name: \"%s\", Environment: \"%s\", Date: \"%v\", Status: \"%s\"\n", entry.name, entry.environment, entry.date, entry.status)
	}
}

type TestEntry struct {
	name        string
	environment string
	date        time.Time
	status      string
}

// Reads CSV `file` and consumes each line to be a single TestEntry.
func ReadData(file *os.File) []TestEntry {
	testEntries := []TestEntry{}

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
				exit(fmt.Sprintf("Data CSV in incorrect format. Expected 6 columns, but got %d", len(fields)), fmt.Errorf("Bad CSV format"))
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
			}
			testEntries = append(testEntries, TestEntry{
				name:        fields[3],
				environment: fields[2],
				date:        date,
				status:      fields[4],
			})
		}
	}
	return testEntries
}

// exit will exit and clean up minikube
func exit(msg string, err error) {
	fmt.Printf("WithError(%s)=%v called from:\n%s", msg, err, debug.Stack())
	os.Exit(60)
}
