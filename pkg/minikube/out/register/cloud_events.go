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

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/glog"
	guuid "github.com/google/uuid"
)

const (
	specVersion = "1.0"
)

var (
	OutputFile io.Writer = os.Stdout
	GetUUID              = randomID
)

func printAsCloudEvent(log Log, data map[string]string) {
	event := cloudevents.NewEvent()
	event.SetSource("https://minikube.sigs.k8s.io/")
	event.SetType(log.Type())
	event.SetSpecVersion(specVersion)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		glog.Warningf("error setting data: %v", err)
	}
	event.SetID(GetUUID())
	json, err := event.MarshalJSON()
	if err != nil {
		glog.Warningf("error marashalling event: %v", err)
	}
	fmt.Fprintln(OutputFile, string(json))
}

func randomID() string {
	return guuid.New().String()
}
