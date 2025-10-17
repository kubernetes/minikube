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
	"slices"
	"strings"

	"golang.org/x/term"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
)

// AskForYesNoConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func AskForYesNoConfirmation(s string, posResponses, negResponses []string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		out.Stringf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		switch r := strings.ToLower(strings.TrimSpace(response)); {
		case slices.Contains(posResponses, r):
			return true
		case slices.Contains(negResponses, r):
			return false
		default:
			out.Err("Please type yes or no:")
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
			out.Err("--Error, please enter a value:")
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
	out.String(s)

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
		)

		if hidden {
			t := term.NewTerminal(readWriter, "")
			response, err = t.ReadPassword(promptString)
		} else {
			t := term.NewTerminal(readWriter, promptString)
			response, err = t.ReadLine()
		}

		if err != nil {
			return "", err
		}
		response = strings.TrimSpace(response)
		if len(response) == 0 {
			out.WarningT("Please enter a value:")
			continue
		}
		return response, nil
	}
}

// AskForPasswordValue asks for a password value, while hiding the input
func AskForPasswordValue(s string) string {

	stdInFd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(stdInFd)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := term.Restore(stdInFd, oldState); err != nil {
			klog.Errorf("terminal restore failed: %v", err)
		}
	}()

	result, err := concealableAskForStaticValue(os.Stdin, s, true)
	if err != nil {
		defer log.Fatal(err)
	}
	return result
}

// AskForStaticValidatedValue asks for a single value to enter and check for valid input
func AskForStaticValidatedValue(s string, validator func(s string) bool) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		response := getStaticValue(reader, s)

		// Can't have zero length
		if len(response) == 0 {
			out.Err("--Error, please enter a value:")
			continue
		}
		if !validator(response) {
			out.Err("--Invalid input, please enter a value:")
			continue
		}
		return response
	}
}
