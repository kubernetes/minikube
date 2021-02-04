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

type Data struct {
	headers []string
	logs    []string
}

// Report is created from the log file.
func Report(lines int) (*Data, error) {
	if lines <= 0 {
		return nil, fmt.Errorf("number of lines must be 1 or greater")
	}
	if currentLogFile == nil {
		if err := setLogFile(); err != nil {
			return nil, fmt.Errorf("failed to set the log file: %v", err)
		}
	}
	var l []string
	s := bufio.NewScanner(currentLogFile)
	for s.Scan() {
		// pop off the earliest line if already at desired log length
		if len(l) == lines {
			l = l[1:]
		}
		l = append(l, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("failed to read from audit file: %v", err)
	}
	r := &Data{
		[]string{"Command", "Args", "Profile", "User", "Start Time", "End Time"},
		l,
	}
	return r, nil
}

// Table creates a formatted table using last n lines from the report.
func (r *Data) Table() (string, error) {
	t, err := linesToTable(r.logs, r.headers)
	if err != nil {
		return "", fmt.Errorf("failed to convert lines to table: %v", err)
	}
	return t, nil
}
