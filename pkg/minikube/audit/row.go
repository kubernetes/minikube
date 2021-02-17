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

// row is the log of a single command.
type row struct {
	args      string
	command   string
	endTime   string
	profile   string
	startTime string
	user      string
	version   string
	Data      map[string]string `json:"data"`
}

// Type returns the cloud events compatible type of this struct.
func (e *row) Type() string {
	return "io.k8s.sigs.minikube.audit"
}

// assignFields converts the map values to their proper fields
func (e *row) assignFields() {
	e.args = e.Data["args"]
	e.command = e.Data["command"]
	e.endTime = e.Data["endTime"]
	e.profile = e.Data["profile"]
	e.startTime = e.Data["startTime"]
	e.user = e.Data["user"]
	e.version = e.Data["version"]
}

// toMap combines fields into a string map
func (e *row) toMap() map[string]string {
	return map[string]string{
		"args":      e.args,
		"command":   e.command,
		"endTime":   e.endTime,
		"profile":   e.profile,
		"startTime": e.startTime,
		"user":      e.user,
		"version":   e.version,
	}
}

// newRow returns a new audit type.
func newRow(command string, args string, user string, version string, startTime time.Time, endTime time.Time, profile ...string) *row {
	p := viper.GetString(config.ProfileName)
	if len(profile) > 0 {
		p = profile[0]
	}
	return &row{
		args:      args,
		command:   command,
		endTime:   endTime.Format(constants.TimeFormat),
		profile:   p,
		startTime: startTime.Format(constants.TimeFormat),
		user:      user,
		version:   version,
	}
}

// toFields converts a row to an array of fields.
func (e *row) toFields() []string {
	return []string{e.command, e.args, e.profile, e.user, e.version, e.startTime, e.endTime}
}

// logsToRows converts audit logs into arrays of rows.
func logsToRows(logs []string) ([]row, error) {
	rows := []row{}
	for _, l := range logs {
		r := row{}
		if err := json.Unmarshal([]byte(l), &r); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %q: %v", l, err)
		}
		r.assignFields()
		rows = append(rows, r)
	}
	return rows, nil
}

// rowsToTable converts audit rows into a formatted table.
func rowsToTable(rows []row, headers []string) string {
	c := [][]string{}
	for _, r := range rows {
		c = append(c, r.toFields())
	}
	b := new(bytes.Buffer)
	t := tablewriter.NewWriter(b)
	t.SetHeader(headers)
	t.SetAutoFormatHeaders(false)
	t.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	t.SetCenterSeparator("|")
	t.AppendBulk(c)
	t.Render()
	return b.String()
}
