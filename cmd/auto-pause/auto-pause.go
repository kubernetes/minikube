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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var unpauseRequests = make(chan struct{})
var done = make(chan struct{})
var mu sync.Mutex

var runtimePaused bool
var version = "0.0.1"

var runtime = flag.String("container-runtime", "docker", "Container runtime to use for (un)pausing")

func main() {
	flag.Parse()

	// TODO: #10595 make this configurable
	const interval = time.Minute * 1

	// Check current state
	alreadyPaused()

	// channel for incoming messages
	go func() {
		for {
			// On each iteration new timer is created
			select {
			// TODO: #10596 make it memory-leak proof
			case <-time.After(interval):
				runPause()
			case <-unpauseRequests:
				fmt.Printf("Got request\n")
				if runtimePaused {
					runUnpause()
				}

				done <- struct{}{}
			}
		}
	}()

	http.HandleFunc("/", handler) // each request calls handler
	fmt.Printf("Starting auto-pause server %s at port 8080 \n", version)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

// handler echoes the Path component of the requested URL.
func handler(w http.ResponseWriter, r *http.Request) {
	unpauseRequests <- struct{}{}
	<-done
	fmt.Fprintf(w, "allow")
}

func runPause() {
	mu.Lock()
	defer mu.Unlock()
	if runtimePaused {
		return
	}

	r := command.NewExecRunner(true)

	cr, err := cruntime.New(cruntime.Config{Type: *runtime, Runner: r})
	if err != nil {
		exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
	}

	uids, err := cluster.Pause(cr, r, []string{"kube-system"})
	if err != nil {
		exit.Error(reason.GuestPause, "Pause", err)
	}

	runtimePaused = true

	out.Step(style.Unpause, "Paused {{.count}} containers", out.V{"count": len(uids)})
}

func runUnpause() {
	fmt.Println("unpausing...")
	mu.Lock()
	defer mu.Unlock()

	r := command.NewExecRunner(true)

	cr, err := cruntime.New(cruntime.Config{Type: *runtime, Runner: r})
	if err != nil {
		exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
	}

	uids, err := cluster.Unpause(cr, r, nil)
	if err != nil {
		exit.Error(reason.GuestUnpause, "Unpause", err)
	}
	runtimePaused = false

	out.Step(style.Unpause, "Unpaused {{.count}} containers", out.V{"count": len(uids)})
}

func alreadyPaused() {
	mu.Lock()
	defer mu.Unlock()

	r := command.NewExecRunner(true)
	cr, err := cruntime.New(cruntime.Config{Type: *runtime, Runner: r})
	if err != nil {
		exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
	}

	runtimePaused, err = cluster.CheckIfPaused(cr, []string{"kube-system"})
	if err != nil {
		exit.Error(reason.GuestCheckPaused, "Fail check if container paused", err)
	}
	out.Step(style.Check, "containers paused status: {{.paused}}", out.V{"paused": runtimePaused})
}
