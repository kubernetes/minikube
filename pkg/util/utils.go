/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package util

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Until endlessly loops the provided function until a message is received on the done channel.
// The function will wait the duration provided in sleep between function calls. Errors will be sent on provider Writer.
func Until(fn func() error, w io.Writer, name string, sleep time.Duration, done <-chan struct{}) {
	var exitErr error
	for {
		select {
		case <-done:
			return
		default:
			exitErr = fn()
			if exitErr == nil {
				fmt.Fprintf(w, Pad("%s: Exited with no errors.\n"), name)
			} else {
				fmt.Fprintf(w, Pad("%s: Exit with error: %v"), name, exitErr)
			}

			// wait provided duration before trying again
			time.Sleep(sleep)
		}
	}
}

func Pad(str string) string {
	return fmt.Sprint("\n%s\n", str)
}

// If the file represented by path exists and
// readable, return true otherwise return false.
func CanReadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	defer f.Close()

	return true
}
