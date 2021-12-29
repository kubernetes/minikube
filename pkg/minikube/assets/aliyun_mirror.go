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

package assets

import (
	"encoding/json"
	"fmt"

	"k8s.io/minikube/deploy/addons"
)

// AliyunMirror list of images from Aliyun mirror
var AliyunMirror = loadAliyunMirror()

func loadAliyunMirror() map[string]string {
	data, err := addons.AliyunMirror.ReadFile("aliyun_mirror.json")
	if err != nil {
		panic(fmt.Sprintf("Failed to load aliyun_mirror.json: %v", err))
	}
	var mirror map[string]string
	err = json.Unmarshal(data, &mirror)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse aliyun_mirror.json: %v", err))
	}
	return mirror
}
