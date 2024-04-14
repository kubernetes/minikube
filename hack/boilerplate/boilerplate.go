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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	boilerplateDir = flag.String("boilerplate-dir", ".", "Directory containing boilerplate files")
	rootDir        = flag.String("root-dir", "../../", "Root directory to examine")
	verbose        = flag.Bool("verbose", false, "Verbose mode")
	skippedPaths   = regexp.MustCompile(`Godeps|third_party|_gopath|_output|\.git|cluster/env.sh|vendor|test/e2e/generated/bindata.go|site/themes/docsy|test/integration/testdata`)
	windowdNewLine = regexp.MustCompile(`\r`)
	txtExtension   = regexp.MustCompile(`\.txt`)
	goBuildTag     = regexp.MustCompile(`(?m)^(//go:build.*\n)+\n`)
	shebang        = regexp.MustCompile(`(?m)^(#!.*\n)\n*`)
	copyright      = regexp.MustCompile(`Copyright YEAR`)
	copyrightReal  = regexp.MustCompile(`Copyright \d{4}`)
)

func main() {
	flag.Parse()

	refs, err := extensionToBoilerplate(*boilerplateDir)
	if err != nil {
		log.Fatal(err)
	}
	if len(refs) == 0 {
		log.Fatal("no references found in ", *boilerplateDir)
	}

	files, err := filesToCheck(*rootDir, refs)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		pass, err := filePasses(file, refs[filepath.Ext(file)])
		if err != nil {
			log.Println(err)
		}
		if !pass {
			path, err := filepath.Abs(file)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(path)
		}
	}
}

func extensionToBoilerplate(dir string) (map[string][]byte, error) {
	refs := make(map[string][]byte)
	files, _ := filepath.Glob(dir + "/*.txt")

	for _, filename := range files {
		extension := strings.ToLower(filepath.Ext(txtExtension.ReplaceAllString(filename, "")))
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		refs[extension] = windowdNewLine.ReplaceAll(data, nil)
	}

	if *verbose {
		dir, err := filepath.Abs(dir)
		if err != nil {
			return refs, err
		}
		fmt.Printf("Found %v boilerplates in %v for the following extensions:", len(refs), dir)
		for ext := range refs {
			fmt.Printf(" %v", ext)
		}
		fmt.Println()
	}

	return refs, nil
}

func filePasses(filename string, expectedBoilerplate []byte) (bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	data = windowdNewLine.ReplaceAll(data, nil)

	extension := filepath.Ext(filename)

	if extension == ".go" {
		data = goBuildTag.ReplaceAll(data, nil)
	}

	if extension == ".sh" {
		data = shebang.ReplaceAll(data, nil)
	}

	if len(data) < len(expectedBoilerplate) {
		return false, nil
	}

	data = data[:len(expectedBoilerplate)]

	if copyright.Match(data) {
		return false, nil
	}

	data = copyrightReal.ReplaceAll(data, []byte(`Copyright YEAR`))

	return bytes.Equal(data, expectedBoilerplate), nil
}

func filesToCheck(rootDir string, extensions map[string][]byte) ([]string, error) {
	var outFiles []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, _ error) error {
		cwd, _ := os.Getwd()
		re := regexp.MustCompile(`\\`)
		re = regexp.MustCompile(`^` + re.ReplaceAllString(cwd, `\\`))
		if !info.IsDir() && !skippedPaths.MatchString(re.ReplaceAllString(filepath.Dir(path), "")) {
			if extensions[strings.ToLower(filepath.Ext(path))] != nil {
				outFiles = append(outFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if *verbose {
		rootDir, err = filepath.Abs(rootDir)
		if err != nil {
			return outFiles, err
		}
		fmt.Printf("Found %v files to check in %v\n\n", len(outFiles), rootDir)
	}
	return outFiles, nil
}
