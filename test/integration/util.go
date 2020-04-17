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

package integration

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// ReadLineWithTimeout reads a line of text from a buffer with a timeout
func ReadLineWithTimeout(b *bufio.Reader, timeout time.Duration) (string, error) {
	s := make(chan string)
	e := make(chan error)
	go func() {
		read, err := b.ReadString('\n')
		if err != nil {
			e <- err
		} else {
			s <- read
		}
		close(s)
		close(e)
	}()

	select {
	case line := <-s:
		return line, nil
	case err := <-e:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout after %s", timeout)
	}
}

// UniqueProfileName returns a reasonably unique profile name
func UniqueProfileName(prefix string) string {
	if *forceProfile != "" {
		return *forceProfile
	}
	if NoneDriver() {
		return "minikube"
	}
	// example: prefix-20200413T162239-3215
	return fmt.Sprintf("%s-%s-%d", prefix, time.Now().Format("20060102150405"), os.Getpid())
}
