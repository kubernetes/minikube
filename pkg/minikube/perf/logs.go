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

package perf

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
	"time"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// timeCommandLogs runs command and watches stdout to time how long each new log takes
// it stores each log, and the time it took, in result
func timeCommandLogs(cmd *exec.Cmd) (*result, error) {
	// matches each log with the amount of time spent on that log
	r := newResult()

	output = strings.ToLower(output)
	if output != "text" && statusFormat != defaultStatusFormat {
		exit.Message(reason.Usage, "Cannot use both --output and --format options")
		}
	out.SetJSON(output == "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "getting stdout pipe")
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)

	log.Printf("Running: %v...", cmd.Args)
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "starting cmd")
	}

	timer := time.Now()
	var logs []string
	var timings []float64

	for scanner.Scan() {
		log := scanner.Text()
		// this is the time it took to complete the previous log
		timeTaken := time.Since(timer).Seconds()
		klog.Infof("%f: %s", timeTaken, log)

		timer = time.Now()
		logs = append(logs, log)
		timings = append(timings, timeTaken)
	}
	// add the time it took to get from the final log to finishing the command
	timings = append(timings, time.Since(timer).Seconds())
	for i, log := range logs {
		r.addTimedLog(strings.Trim(log, "\n"), timings[i+1])
	}

	if err := cmd.Wait(); err != nil {
		return nil, errors.Wrap(err, "waiting for minikube")
	}
	return r, nil
}
