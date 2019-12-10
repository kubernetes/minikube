/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package localpath

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

// MinikubeHome is the name of the minikube home directory environment variable.
const MinikubeHome = "MINIKUBE_HOME"

// ConfigFile is the path of the config file
var ConfigFile = MakeMiniPath("config", "config.json")

// MiniPath returns the path to the user's minikube dir
func MiniPath() string {
	if os.Getenv(MinikubeHome) == "" {
		return filepath.Join(homedir.HomeDir(), ".minikube")
	}
	if filepath.Base(os.Getenv(MinikubeHome)) == ".minikube" {
		return os.Getenv(MinikubeHome)
	}
	return filepath.Join(os.Getenv(MinikubeHome), ".minikube")
}

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{MiniPath()}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// MachinePath returns the Minikube machine path of a machine
func MachinePath(machine string, miniHome ...string) string {
	miniPath := MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}
	return filepath.Join(miniPath, "machines", machine)
}
