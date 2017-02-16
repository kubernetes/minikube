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

package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// AskForYesNoConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func AskForYesNoConfirmation(s string, posResponses, negResponses []string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if containsString(posResponses, response) {
			return true
		} else if containsString(negResponses, response) {
			return false
		} else {
			fmt.Println("Please type yes or no:")
			return AskForYesNoConfirmation(s, posResponses, negResponses)
		}
	}
}

// AskForStaticValue asks for a single value to enter
func AskForStaticValue(s string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.TrimSpace(response)

		// Can't have zero length
		if len(response) == 0 {
			fmt.Println("--Error, please enter a value:")
			return AskForStaticValue(s)
		}
		return response
	}
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}
