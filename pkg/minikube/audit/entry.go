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
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// entry represents the execution of a command.
type entry struct {
	Data map[string]string `json:"data"`
}

// Type returns the cloud events compatible type of this struct.
func (e *entry) Type() string {
	return "io.k8s.sigs.minikube.audit"
}

// newEntry returns a new audit type.
func newEntry(command string, args string, user string, startTime time.Time, endTime time.Time) *entry {
	return &entry{
		map[string]string{
			"args":      args,
			"command":   command,
			"endTime":   endTime.Format(constants.TimeFormat),
			"profile":   viper.GetString(config.ProfileName),
			"startTime": startTime.Format(constants.TimeFormat),
			"user":      user,
		},
	}
}

// toFields converts an entry to an array of fields.
func (e *entry) toFields() []string {
	d := e.Data
	return []string{d["command"], d["args"], d["profile"], d["user"], d["startTime"], d["endTime"]}
}

// linesToFields converts audit lines into arrays of fields.
func linesToFields(lines []string) ([][]string, error) {
	c := [][]string{}
	for _, l := range lines {
		e := &entry{}
		if err := json.Unmarshal([]byte(l), e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %q: %v", l, err)
		}
		c = append(c, e.toFields())
	}
	return c, nil
}

// linesToTable converts audit lines into a formatted table.
func linesToTable(lines []string, headers []string) (string, error) {
	f, err := linesToFields(lines)
	if err != nil {
		return "", fmt.Errorf("failed to convert lines to fields: %v", err)
	}
	b := new(bytes.Buffer)
	t := tablewriter.NewWriter(b)
	t.SetHeader(headers)
	t.SetAutoFormatHeaders(false)
	t.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	t.SetCenterSeparator("|")
	t.AppendBulk(f)
	t.Render()
	return b.String(), nil
}
