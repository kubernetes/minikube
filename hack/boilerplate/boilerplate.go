/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var skippedDirs = regexp.MustCompile(`Godeps|third_party|_gopath|_output|\.git|cluster/env.sh|vendor|test/e2e/generated/bindata.go`)

func main() {
	rootdir, _ := os.Getwd()
	rootdir += "/../../"
	var boilerplateDir string
	var filenames []string

	for i := 1; i < len(os.Args); i++ {
		re := regexp.MustCompile("^--.*")
		if re.MatchString(os.Args[i]) {
			if os.Args[i] == "--rootdir" {
				rootdir = os.Args[i+1]
				i++
				continue
			}
			if os.Args[i] == "--boilerplate-dir" {
				boilerplateDir = os.Args[i+1]
				i++
				continue
			}
		} else {
			filenames = append(filenames, os.Args[i])
		}

	}

	if len(boilerplateDir) == 0 {
		boilerplateDir = filepath.Join(rootdir, "hack/boilerplate")
	}

	refs := getRefs(boilerplateDir)
	files := getFileList(rootdir, refs, filenames)
	for _, file := range files {
		if !filePasses(file, refs[getFileExtension(file)]) {
			fmt.Println(file)
		}
	}

}

func getRefs(dir string) map[string][]byte {
	refs := make(map[string][]byte)
	files, _ := filepath.Glob(dir + "/*.txt")
	for _, filename := range files {
		extension := strings.ToLower(strings.Split(filename, ".")[1])
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		re := regexp.MustCompile(`\r`)
		refs[extension] = re.ReplaceAll(data, nil)
	}
	return refs
}

func filePasses(filename string, ref []byte) bool {
	var re *regexp.Regexp
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	re = regexp.MustCompile(`\r`)
	data = re.ReplaceAll(data, nil)

	extension := getFileExtension(filename)

	// remove build tags from the top of Go files
	if extension == "go" {
		// \r is necessary for windows file endings
		re = regexp.MustCompile(`(?m)^(// \+build.*\r{0,1}\n)+\r{0,1}\n`)
		data = re.ReplaceAll(data, nil)
	}

	// remove shebang from the top of shell files
	if extension == "sh" {
		// \r is necessary for windows file endings
		// re := regexp.MustCompile(`(?m)^(// \+build.*\r{0,1}\n)+\r{0,1}\n`)
		re = regexp.MustCompile(`(?m)^(#!.*\r{0,1}\n)(\r{0,1}\n)*`)
		data = re.ReplaceAll(data, nil)
	}

	// if our test file is smaller than the reference it surely fails!
	if len(data) < len(ref) {
		return false
	}

	data = data[:len(ref)]

	// Search for "Copyright YEAR" which exists in the boilerplate, but shouldn't in the real thing
	re = regexp.MustCompile(`Copyright YEAR`)
	if re.Match(data) {
		return false
	}

	// Replace all occurrences of the regex "Copyright \d{4}" with "Copyright YEAR"
	re = regexp.MustCompile(`Copyright \d{4}`)
	data = re.ReplaceAll(data, []byte(`Copyright YEAR`))

	return bytes.Equal(data, ref)
}

// get the file extensin or the filename if the file has no extension
func getFileExtension(filename string) string {
	splitted := strings.Split(filepath.Base(filename), ".")
	return strings.ToLower(splitted[len(splitted)-1])
}

func getFileList(rootDir string, extensions map[string][]byte, files []string) []string {
	var outFiles []string
	if len(files) == 0 {
		err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
			// println(path)
			if !info.IsDir() && !skippedDirs.MatchString(filepath.Dir(path)) {
				if extensions[strings.ToLower(getFileExtension(path))] != nil {
					outFiles = append(outFiles, path)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		outFiles = files
	}
	return outFiles
}
