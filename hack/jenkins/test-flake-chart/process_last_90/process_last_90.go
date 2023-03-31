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
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	dataFile := flag.String("source", "", "The source of the csv file to process")
	dataLast90File := flag.String("target", "", "The target of the csv file containing last 90 days of data")
	flag.Parse()

	if *dataFile == "" || *dataLast90File == "" {
		fmt.Println("All flags are required and cannot be empty")
		flag.PrintDefaults()
		os.Exit(1)
	}

	data, err := os.Open(*dataFile)
	if err != nil {
		log.Fatalf("failed opening source file %q: %v", *dataFile, err)
	}

	dataLast90, err := os.OpenFile(*dataLast90File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed creating target file %q: %v", *dataLast90File, err)
	}

	dw := bufio.NewWriter(dataLast90)

	cutoffDate := time.Now().AddDate(0, 0, -90)

	// to save space we don't repeat duplicate back to back lines if they have the same value
	// for example:
	// 2021-09-10,TestOffline,Passed
	// ,TestForceSystemd,
	// So the date for line two will be empty, so we need to remember the last line that had a date if it's within last 90 days or not
	validDate := true

	s := bufio.NewScanner(data)
	for s.Scan() {
		line := s.Text()
		stringDate := strings.Split(line, ",")[1]

		// copy headers
		if stringDate == "Test Date" {
			write(dw, line)
			continue
		}

		if stringDate == "" {
			if validDate {
				write(dw, line)
			}
			continue
		}

		testDate, err := time.Parse("2006-01-02", stringDate)
		if err != nil {
			log.Fatalf("failed to parse date %q: %v", stringDate, err)
		}
		if testDate.Before(cutoffDate) {
			validDate = false
			continue
		}
		validDate = true
		write(dw, line)
	}
	if err := s.Err(); err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	if err := dw.Flush(); err != nil {
		log.Fatalf("failed to flush data writer: %v", err)
	}
	if err := data.Close(); err != nil {
		log.Fatalf("failed to close source file: %v", err)
	}
	if err := dataLast90.Close(); err != nil {
		log.Fatalf("failed to close target file: %v", err)
	}
}

func write(dw *bufio.Writer, line string) {
	if _, err := dw.WriteString(line + "\n"); err != nil {
		log.Fatalf("failed to write to data writer: %v", err)
	}
}
