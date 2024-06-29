/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
)

const (
	maxItemEnv = 10
)

// This program requires three arguments
// $1 is the Pr Number
// $2 is the ROOT_JOB
// $3 is the file containing a list of finished environments, one item per line
func main() {
	ctx := context.Background()
	client, err := storage.NewClient(context.Background())
	if err != nil {
		fmt.Printf("failed to connect to gcp: %v\n", err)
		os.Exit(1)
	}
	// parse the command line arguments
	if len(os.Args) != 4 {
		fmt.Println("Wrong number of arguments. Usage: go run . <PR number> <Root job id> <environment list file>")

		os.Exit(1)
	}
	pr := os.Args[1]
	rootJob := os.Args[2]
	// read the environment names
	envList, err := parseEnvironmentList(os.Args[3])
	if err != nil {
		fmt.Printf("failed to read %s, err: %v\n", os.Args[3], err)
		os.Exit(1)
	}
	// fetch the test results
	testSummaries, err := testSummariesFromGCP(ctx, pr, rootJob, envList, client)
	if err != nil {
		fmt.Printf("failed to load summaries: %v\n", err)
		os.Exit(1)
	}
	// fetch the pre-calculated flake rates
	flakeRates, err := flakeRate(ctx, client)
	if err != nil {
		fmt.Printf("failed to load flake rates: %v\n", err)
		os.Exit(1)
	}
	// generate and send the message
	msg := generateCommentMessage(testSummaries, flakeRates, pr, rootJob)
	fmt.Println(msg)
}
