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

package register

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/glog"
	guuid "github.com/google/uuid"
)

const (
	specVersion = "1.0"
)

var (
	outputFile io.Writer = os.Stdout
	GetUUID              = randomID

	eventFile *os.File
)

// SetOutputFile sets the writer to emit all events to
func SetOutputFile(w io.Writer) {
	outputFile = w
}

// SetEventLogPath sets the path of an event log file
func SetEventLogPath(path string) {
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
			glog.Errorf("Error creating profile directory: %v", err)
			return
		}
	}

	f, err := os.Create(path)
	if err != nil {
		glog.Errorf("unable to write to %s: %v", path, err)
		return
	}
	eventFile = f
}

// cloudEvent creates a CloudEvent from a log object & associated data
func cloudEvent(log Log, data map[string]string) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource("https://minikube.sigs.k8s.io/")
	event.SetType(log.Type())
	event.SetSpecVersion(specVersion)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		glog.Warningf("error setting data: %v", err)
	}
	event.SetID(GetUUID())
	return event
}

// print JSON output to configured writer
func printAsCloudEvent(log Log, data map[string]string) {
	event := cloudEvent(log, data)

	bs, err := event.MarshalJSON()
	if err != nil {
		glog.Errorf("error marshalling event: %v", err)
		return
	}
	fmt.Fprintln(outputFile, string(bs))
}

// print JSON output to configured writer, and record it to disk
func printAndRecordCloudEvent(log Log, data map[string]string) {
	event := cloudEvent(log, data)

	bs, err := event.MarshalJSON()
	if err != nil {
		glog.Errorf("error marshalling event: %v", err)
		return
	}
	fmt.Fprintln(outputFile, string(bs))

	if eventFile != nil {
		go storeEvent(bs)
	}
}

func storeEvent(bs []byte) {
	fmt.Fprintln(eventFile, string(bs))
	if err := eventFile.Sync(); err != nil {
		glog.Warningf("even file flush failed: %v", err)
	}
}

// record cloud event to disk
func recordCloudEvent(log Log, data map[string]string) {
	if eventFile == nil {
		return
	}

	go func() {
		event := cloudEvent(log, data)
		bs, err := event.MarshalJSON()
		if err != nil {
			glog.Errorf("error marshalling event: %v", err)
			return
		}
		storeEvent(bs)
	}()
}

func randomID() string {
	return guuid.New().String()
}
