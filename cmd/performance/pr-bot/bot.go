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
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/perf/monitor"
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
	client := monitor.NewClient(ctx, monitor.GithubOwner, monitor.GithubRepo)
	prs, err := client.ListOpenPRsWithLabel(monitor.OkToTestLabel)
	if err != nil {
		return errors.Wrap(err, "listing open prs")
	}
	log.Print("got prs:", prs)
	for _, pr := range prs {
		log.Printf("~~~ Analyzing PR %d ~~~", pr)
		newCommitsExist, err := client.NewCommitsExist(pr, monitor.BotName)
		if err != nil {
			return err
		}
		if !newCommitsExist {
			log.Println("New commits don't exist, skipping rerun...")
			continue
		}
		var message string
		message, err = monitor.RunMkcmp(ctx, pr)
		if err != nil {
			message = fmt.Sprintf("Error: %v\n%s", err, message)
		}
		log.Printf("message for pr %d:\n%s\n", pr, message)
		if err := client.CommentOnPR(pr, message); err != nil {
			return err
		}
		log.Print("successfully commented on PR")
	}
	return nil
}
