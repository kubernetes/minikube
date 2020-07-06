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

package out

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func printAsCloudEvent(eventType string, data map[string]string) {
	event := cloudevents.NewEvent()
	event.SetSource("https://minikube.sigs.k8s.io/")
	event.SetType(eventType)
	event.SetData(cloudevents.ApplicationJSON, data)
	json, err := event.MarshalJSON()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(json))
}
