/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"os"
	"os/signal"
	"syscall"
)

// This is used to unittest functions that kill processes,
// in a cross-platform way.
func main() {
	ch := make(chan os.Signal, 1)
	done := make(chan struct{})
	defer close(ch)

	signal.Notify(ch, syscall.SIGHUP)
	defer signal.Stop(ch)

	go func() {
		<-ch
		close(done)
	}()

	<-done
}
