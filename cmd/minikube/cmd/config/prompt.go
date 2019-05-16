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
	"io"
	"log"
	"os"
	"strings"

	"github.com/golang/glog"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/minikube/pkg/minikube/console"
)

// AskForYesNoConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func AskForYesNoConfirmation(s string, posResponses, negResponses []string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		console.Out("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		switch r := strings.ToLower(strings.TrimSpace(response)); {
		case containsString(posResponses, r):
			return true
		case containsString(negResponses, r):
			return false
		default:
			console.Err("Please type yes or no:")
		}
	}
}

// AskForStaticValue asks for a single value to enter
func AskForStaticValue(s string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		response := getStaticValue(reader, s)

		// Can't have zero length
		if len(response) == 0 {
			console.Err("--Error, please enter a value:")
			continue
		}
		return response
	}
}

// AskForStaticValueOptional asks for a optional single value to enter, can just skip enter
func AskForStaticValueOptional(s string) string {
	reader := bufio.NewReader(os.Stdin)

	return getStaticValue(reader, s)
}

func getStaticValue(reader *bufio.Reader, s string) string {
	console.Out("%s", s)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	response = strings.TrimSpace(response)
	return response
}

func concealableAskForStaticValue(readWriter io.ReadWriter, promptString string, hidden bool) (string, error) {
	for {
		var (
			response string
			err      error
			term     *terminal.Terminal
		)

		if hidden {
			term = terminal.NewTerminal(readWriter, "")
			response, err = term.ReadPassword(promptString)
		} else {
			term = terminal.NewTerminal(readWriter, promptString)
			response, err = term.ReadLine()
		}

		if err != nil {
			return "", err
		}
		response = strings.TrimSpace(response)
		if len(response) == 0 {
			console.Warning("Please enter a value:")
			continue
		}
		return response, nil
	}
}

// AskForPasswordValue asks for a password value, while hiding the input
func AskForPasswordValue(s string) string {

	stdInFd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(stdInFd)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := terminal.Restore(stdInFd, oldState); err != nil {
			glog.Errorf("terminal restore failed: %v", err)
		}
	}()

	result, err := concealableAskForStaticValue(os.Stdin, s, true)
	if err != nil {
		log.Fatal(err)
	}
	return result
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
