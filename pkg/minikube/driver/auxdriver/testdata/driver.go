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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "v1.2.3"
const commit = "1af8bdc072232de4b1fec3b6cc0e8337e118bc83"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "no command specified\n")
		os.Exit(1)
	}

	if os.Args[1] != "version" {
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(1)
	}

	// We use a single executable to emulate different driver outputs.
	driverPersonality := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")

	switch driverPersonality {
	case "valid":
		fmt.Printf("version: %s\n", version)
		fmt.Printf("commit: %s\n", commit)
	case "no-version":
		fmt.Printf("commit: %s\n", commit)
	case "no-commit":
		fmt.Printf("version: %s\n", version)
	case "invalid":
		fmt.Println("invalid yaml")
	case "fail":
		fmt.Fprintf(os.Stderr, "no version for you!\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown personality %q\n", driverPersonality)
		os.Exit(1)
	}
}
