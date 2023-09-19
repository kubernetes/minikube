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

package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/version"
)

// userName pulls the user flag, if empty gets the os username.
func userName() string {
	u := viper.GetString(config.UserFlag)
	if u != "" {
		return u
	}
	osUser, err := user.Current()
	if err != nil {
		return "UNKNOWN"
	}
	return osUser.Username
}

// args concats the args into space delimited string.
func args() string {
	// first arg is binary and second is command, anything beyond is a minikube arg
	if len(os.Args) < 3 {
		return ""
	}
	return strings.Join(os.Args[2:], " ")
}

// Log details about the executed command.
func LogCommandStart() (string, error) {
	if !shouldLog() {
		return "", nil
	}
	id := uuid.New().String()
	r := newRow(pflag.Arg(0), args(), userName(), version.GetVersion(), time.Now(), id)
	if err := appendToLog(r); err != nil {
		return "", err
	}
	return r.id, nil
}

func LogCommandEnd(id string) error {
	if id == "" {
		return nil
	}
	if err := openAuditLog(); err != nil {
		return err
	}
	defer closeAuditLog()
	var logs []string
	s := bufio.NewScanner(currentLogFile)
	for s.Scan() {
		logs = append(logs, s.Text())
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("failed to read from audit file: %v", err)
	}
	closeAuditLog()
	rowSlice, err := logsToRows(logs)
	if err != nil {
		return fmt.Errorf("failed to convert logs to rows: %v", err)
	}
	// have to truncate the audit log while closed as Windows can't truncate an open file
	if err := truncateAuditLog(); err != nil {
		return fmt.Errorf("failed to truncate audit log: %v", err)
	}
	if err := openAuditLog(); err != nil {
		return err
	}
	var entriesNeedsToUpdate int

	startIndex := getStartIndex(len(rowSlice))
	rowSlice = rowSlice[startIndex:]
	for _, v := range rowSlice {
		if v.id == id {
			v.endTime = time.Now().Format(constants.TimeFormat)
			v.Data = v.toMap()
			entriesNeedsToUpdate++
		}
		auditLog, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if _, err = currentLogFile.WriteString(string(auditLog) + "\n"); err != nil {
			return fmt.Errorf("failed to write to audit log: %v", err)
		}
	}
	if entriesNeedsToUpdate == 0 {
		return fmt.Errorf("failed to find a log row with id equals to %v", id)
	}
	return nil
}

func getStartIndex(entryCount int) int {
	// default to 1000 entries
	maxEntries := 1000
	if viper.IsSet(config.MaxAuditEntries) {
		maxEntries = viper.GetInt(config.MaxAuditEntries)
	}
	startIndex := entryCount - maxEntries
	if maxEntries <= 0 || startIndex <= 0 {
		return 0
	}
	return startIndex
}

// shouldLog returns if the command should be logged.
func shouldLog() bool {
	if viper.GetBool(config.SkipAuditFlag) {
		return false
	}

	// in rare chance we get here without a command, don't log
	if pflag.NArg() == 0 {
		return false
	}

	if isDeletePurge() {
		return false
	}

	// commands that should not be logged.
	no := []string{"status", "version", "logs", "generate-docs", "profile"}
	a := pflag.Arg(0)
	for _, c := range no {
		if a == c {
			return false
		}
	}
	return true
}

// isDeletePurge return true if command is delete with purge flag.
func isDeletePurge() bool {
	return pflag.Arg(0) == "delete" && viper.GetBool("purge")
}
