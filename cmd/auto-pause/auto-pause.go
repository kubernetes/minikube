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
	"time"
)

var ticker *time.Ticker
var minutesToPause int

func init() {
	ticker = time.NewTicker(1 * time.Second)
	minutesToPause = 10
	go schedulePause()

}
func main() {
	http.HandleFunc("/", handler) // each request calls handler
	fmt.Printf("Starting server at port 0.0.0.0:8000\n")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

// handler echoes the Path component of the requested URL.
func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Receive request uri %s at port 8000\n", r.RequestURI)
	unPauseIfPaused()
	// reset timer
	fmt.Println("reseting pause counter to another 10")
	minutesToPause = 5
	go schedulePause()
	fmt.Fprintf(w, "allow")
}

func schedulePause() {
	fmt.Println("scheduling pausing ...")
	for minutesToPause > 0 {
		minutesToPause = minutesToPause - 1
		t := <-ticker.C
		fmt.Println("ticking ..", t)
	}
	fmt.Println("Doing Pause")
	pause()
}

func unPauseIfPaused() {
	fmt.Println("unpausing...")

}

func pause() {
	fmt.Println("inside pause")
}
