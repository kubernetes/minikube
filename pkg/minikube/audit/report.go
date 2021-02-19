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

package audit

import (
	"bufio"
	"fmt"
)

// RawReport contains the information required to generate formatted reports.
type RawReport struct {
	headers []string
	rows    []row
}

// Report is created using the last n lines from the log file.
func Report(lastNLines int) (*RawReport, error) {
	if lastNLines <= 0 {
		return nil, fmt.Errorf("last n lines must be 1 or greater")
	}
	if currentLogFile == nil {
		if err := setLogFile(); err != nil {
			return nil, fmt.Errorf("failed to set the log file: %v", err)
		}
	}
	var logs []string
	s := bufio.NewScanner(currentLogFile)
	for s.Scan() {
		// pop off the earliest line if already at desired log length
		if len(logs) == lastNLines {
			logs = logs[1:]
		}
		logs = append(logs, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("failed to read from audit file: %v", err)
	}
	rows, err := logsToRows(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert logs to rows: %v", err)
	}
	r := &RawReport{
		[]string{"Command", "Args", "Profile", "User", "Version", "Start Time", "End Time"},
		rows,
	}
	return r, nil
}

// ASCIITable creates a formatted table using the headers and rows from the report.
func (rr *RawReport) ASCIITable() string {
	return rowsToASCIITable(rr.rows, rr.headers)
}
