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

var incomeCh = make(chan struct{})
var done = make(chan struct{})
var mu sync.Mutex

// TODO: intialize with current state (handle the case that user enables auto-pause after it is already paused)
var runtimePaused = false
var version = "0.0.1"

// TODO: make this configurable to support containerd/cri-o
var runtime = "docker"

func main() {
	// TODO: make this configurable
	const interval = time.Minute * 1
	// channel for incoming messages
	go func() {
		for {
			// On each iteration new timer is created
			select {
			case <-time.After(interval):
				runPause()
			case <-incomeCh:
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
	incomeCh <- struct{}{}
	<-done
	fmt.Fprintf(w, "allow")
}

func runPause() {
	mu.Lock()
	defer mu.Unlock()
	if runtimePaused {
		return
	}

	ids := []string{}

	r := command.NewExecRunner(true)

	cr, err := cruntime.New(cruntime.Config{Type: runtime, Runner: r})
	if err != nil {
		exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
	}

	uids, err := cluster.Pause(cr, r, []string{"kube-system"})
	if err != nil {
		exit.Error(reason.GuestPause, "Pause", err)
	}

	runtimePaused = true
	ids = append(ids, uids...)

	out.Step(style.Unpause, "Paused {{.count}} containers", out.V{"count": len(ids)})
}

func runUnpause() {
	fmt.Println("unpausing...")
	mu.Lock()
	defer mu.Unlock()

	ids := []string{}

	r := command.NewExecRunner(true)

	cr, err := cruntime.New(cruntime.Config{Type: runtime, Runner: r})
	if err != nil {
		exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
	}

	uids, err := cluster.Unpause(cr, r, nil)
	if err != nil {
		exit.Error(reason.GuestUnpause, "Unpause", err)
	}
	ids = append(ids, uids...)
	runtimePaused = false

	out.Step(style.Unpause, "Unpaused {{.count}} containers", out.V{"count": len(ids)})
}
