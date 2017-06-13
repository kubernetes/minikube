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

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	flag "github.com/spf13/pflag"
)

var (
	godepsFile     = flag.String("godeps-file", "", "absolute path to Godeps.json")
	kubernetesPath = flag.String("kubernetes-dir", "", "absolute path to the kubernetes folder")
)

type Dependency struct {
	ImportPath string
	Comment    string `json:",omitempty"`
	Rev        string
}

type Godeps struct {
	ImportPath   string
	GoVersion    string
	GodepVersion string
	Packages     []string `json:",omitempty"` // Arguments to save, if any.
	Deps         []Dependency
}

func main() {
	flag.Parse()
	var g Godeps
	if godepsFile == nil || kubernetesPath == nil {
		log.Fatalf("absolute path to Godeps.json is required")
	}
	f, err := os.OpenFile(*godepsFile, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("cannot open file %q: %v", *godepsFile, err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&g)
	if err != nil {
		log.Fatalf("Unable to parse %q: %v", *godepsFile, err)
	}

	k8sStagingDir := path.Join(*kubernetesPath, "staging", "src", "k8s.io")
	stagedRepos, err := ioutil.ReadDir(k8sStagingDir)
	if err != nil {
		log.Fatalf("Couldn't read kubernetes staging repo: %v", err)
	}

	i := 0
	for _, dep := range g.Deps {
		ignored := false
		for _, stagedRepo := range stagedRepos {
			importPrefix := path.Join("k8s.io", stagedRepo.Name())
			if strings.HasPrefix(dep.ImportPath, importPrefix) {
				ignored = true
			}
		}
		if ignored {
			continue
		}
		g.Deps[i] = dep
		i++
	}
	g.Deps = g.Deps[:i]
	b, err := json.MarshalIndent(g, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	n, err := f.WriteAt(append(b, '\n'), 0)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Truncate(int64(n)); err != nil {
		log.Fatal(err)
	}
}
