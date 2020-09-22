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
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
)

func main() {
	for {
		log.Print("~~~~~~~~~ Starting performance analysis ~~~~~~~~~~~~~~")
		if err := analyzePerformance(context.Background()); err != nil {
			log.Printf("error executing performance analysis: %v", err)
		}
		time.Sleep(10 * time.Minute)
	}
}

// analyzePerformance is responsible for:
//   1. collecting PRs to run performance analysis on
//   2. running mkcmp against those PRs
//   3. commenting results on those PRs
func analyzePerformance(ctx context.Context) error {
	logsFile := "/home/performance-monitor/logs.txt"
	if _, err := os.Stat(logsFile); err != nil {
		return err
	}
	client := monitor.NewClient(context.Background(), "kubernetes", "minikube")
	prs, err := client.ListOpenPRsWithLabel("")
	if err != nil {
		return errors.Wrap(err, "listing open prs")
	}
	log.Print("got prs:", prs)
	// TODO: priyawadhwa@ for each PR we should comment the error if we get one?
	for _, pr := range prs {
		log.Printf("~~~ Analyzing PR %d ~~~", pr)
		newCommitsExist, err := client.NewCommitsExist(pr, "minikube-pr-bot")
		if err != nil {
			return err
		}
		if !newCommitsExist {
			log.Println("New commits don't exist, skipping rerun...")
			continue
		}
		// TODO: priyawadhwa@ we should download mkcmp for each run?
		var message string
		message, err = monitor.RunMkcmp(ctx, pr)
		if err != nil {
			message = fmt.Sprintf("Error: %v\n%s", err, message)
		}
		log.Printf("got message for pr %d:\n%s\n", pr, message)
		if err := client.CommentOnPR(pr, message); err != nil {
			return err
		}
		log.Print("successfully commented on PR")
	}
	return nil
}
