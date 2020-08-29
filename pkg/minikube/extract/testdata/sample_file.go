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

package extract

import "fmt"

func DoSomeStuff() {
	// Test with a URL
	PrintToScreenNoInterface("http://kubernetes.io")

	// Test with something that Go thinks looks like a URL
	PrintToScreenNoInterface("Hint: This is not a URL, come on.")

	// Try with an integer
	PrintToScreenNoInterface("5")

	// Try with a sudo command
	PrintToScreenNoInterface("sudo ls .")

	DoSomeOtherStuff(true, 4, "I think this should work")

	v := "This is a variable with a string assigned"
	PrintToScreenNoInterface(v)
}

func DoSomeOtherStuff(choice bool, i int, s string) {
	// Let's try an if statement
	if choice {
		PrintToScreen("This was a choice: %s", s)
	} else if i > 5 {
		PrintToScreen("Wow another string: %s", i)
	} else {
		// Also try a loop
		for i > 10 {
			PrintToScreenNoInterface("Holy cow I'm in a loop!")
			i = i + 1
		}
	}
}

func PrintToScreenNoInterface(s string) {
	PrintToScreen(s, nil)
}

// This will be the function we'll focus the extractor on
func PrintToScreen(s string, i interface{}) {
	fmt.Printf(s, i)
}
