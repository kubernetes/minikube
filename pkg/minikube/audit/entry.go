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
	"k8s.io/klog"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// singleEntry is the log entry of a single command.
type singleEntry struct {
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
func (e *singleEntry) Type() string {
	return "io.k8s.sigs.minikube.audit"
}

// assignFields converts the map values to their proper fields
func (e *singleEntry) assignFields() {
	e.args = e.Data["args"]
	e.command = e.Data["command"]
	e.endTime = e.Data["endTime"]
	e.profile = e.Data["profile"]
	e.startTime = e.Data["startTime"]
	e.user = e.Data["user"]
	e.version = e.Data["version"]
}

// toMap combines fields into a string map
func (e *singleEntry) toMap() map[string]string {
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

// newEntry returns a new audit type.
func newEntry(command string, args string, user string, version string, startTime time.Time, endTime time.Time) *singleEntry {
	return &singleEntry{
		args:      args,
		command:   command,
		endTime:   endTime.Format(constants.TimeFormat),
		profile:   viper.GetString(config.ProfileName),
		startTime: startTime.Format(constants.TimeFormat),
		user:      user,
		version:   version,
	}
}

// toFields converts an entry to an array of fields.
func (e *singleEntry) toFields() []string {
	return []string{e.command, e.args, e.profile, e.user, e.version, e.startTime, e.endTime}
}

// logsToEntries converts audit logs into arrays of entries.
func logsToEntries(logs []string) ([]singleEntry, error) {
	c := []singleEntry{}
	for _, l := range logs {
		e := singleEntry{}
		if err := json.Unmarshal([]byte(l), &e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %q: %v", l, err)
		}
		e.assignFields()
		c = append(c, e)
	}
	return c, nil
}

// entriesToTable converts audit lines into a formatted table.
func entriesToTable(entries []singleEntry, headers []string) (string, error) {
	c := [][]string{}
	for _, e := range entries {
		c = append(c, e.toFields())
	}
	klog.Info(c)
	b := new(bytes.Buffer)
	t := tablewriter.NewWriter(b)
	t.SetHeader(headers)
	t.SetAutoFormatHeaders(false)
	t.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	t.SetCenterSeparator("|")
	t.AppendBulk(c)
	t.Render()
	return b.String(), nil
}
