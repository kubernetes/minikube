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
	"os"
	"os/exec"
	"os/user"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestAudit(t *testing.T) {
	defer func() { auditOverrideFilename = "" }()

	t.Run("setup", func(t *testing.T) {
		f, err := os.CreateTemp("", "audit.json")
		if err != nil {
			t.Fatalf("failed creating temporary file: %v", err)
		}
		defer f.Close()
		auditOverrideFilename = f.Name()

		s := `{"data":{"args":"-p mini1","command":"start","endTime":"Wed, 03 Feb 2021 15:33:05 MST","profile":"mini1","startTime":"Wed, 03 Feb 2021 15:30:33 MST","user":"user1"},"datacontenttype":"application/json","id":"9b7593cb-fbec-49e5-a3ce-bdc2d0bfb208","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}
{"data":{"args":"-p mini1","command":"start","endTime":"Wed, 03 Feb 2021 15:33:05 MST","profile":"mini1","startTime":"Wed, 03 Feb 2021 15:30:33 MST","user":"user1"},"datacontenttype":"application/json","id":"9b7593cb-fbec-49e5-a3ce-bdc2d0bfb208","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}
{"data":{"args":"--user user2","command":"logs","endTime":"Tue, 02 Feb 2021 16:46:20 MST","profile":"minikube","startTime":"Tue, 02 Feb 2021 16:46:00 MST","user":"user2"},"datacontenttype":"application/json","id":"fec03227-2484-48b6-880a-88fd010b5efd","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}
{"data":{"args":"-p mini1","command":"start","endTime":"Wed, 03 Feb 2021 15:33:05 MST","profile":"mini1","startTime":"Wed, 03 Feb 2021 15:30:33 MST","user":"user1"},"datacontenttype":"application/json","id":"9b7593cb-fbec-49e5-a3ce-bdc2d0bfb208","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}
{"data":{"args":"--user user2","command":"logs","endTime":"Tue, 02 Feb 2021 16:46:20 MST","profile":"minikube","startTime":"Tue, 02 Feb 2021 16:46:00 MST","user":"user2"},"datacontenttype":"application/json","id":"fec03227-2484-48b6-880a-88fd010b5efd","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}
`

		if _, err := f.WriteString(s); err != nil {
			t.Fatalf("failed writing to file: %v", err)
		}
	})

	defer os.Remove(auditOverrideFilename)

	t.Run("username", func(t *testing.T) {
		u, err := user.Current()
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			userFlag string
			want     string
		}{
			{
				"testUser",
				"testUser",
			},
			{
				"",
				u.Username,
			},
		}

		for _, test := range tests {
			viper.Set(config.UserFlag, test.userFlag)

			got := userName()

			if got != test.want {
				t.Errorf("userFlag = %q; username() = %q; want %q", test.userFlag, got, test.want)
			}
		}
	})

	t.Run("args", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		tests := []struct {
			args []string
			want string
		}{
			{
				[]string{"minikube", "start"},
				"",
			},
			{
				[]string{"minikube", "start", "--user", "testUser"},
				"--user testUser",
			},
		}

		for _, test := range tests {
			os.Args = test.args

			got := args()

			if got != test.want {
				t.Errorf("os.Args = %q; args() = %q; want %q", os.Args, got, test.want)
			}
		}
	})

	t.Run("shouldLog", func(t *testing.T) {
		oldCommandLine := pflag.CommandLine
		defer func() {
			pflag.CommandLine = oldCommandLine
			pflag.Parse()
		}()

		tests := []struct {
			args []string
			want bool
		}{
			{
				[]string{"minikube", "start"},
				true,
			},
			{
				[]string{"minikube", "delete"},
				true,
			},
			{
				[]string{"minikube", "status"},
				false,
			},
			{
				[]string{"minikube", "version"},
				false,
			},
			{
				[]string{"minikube"},
				false,
			},
			{
				[]string{"minikube", "delete", "--purge"},
				false,
			},
		}

		for _, test := range tests {
			mockArgs(t, test.args)

			got := shouldLog()

			if got != test.want {
				t.Errorf("test.args = %q; shouldLog() = %t; want %t", test.args, got, test.want)
			}
		}
	})

	t.Run("isDeletePurge", func(t *testing.T) {
		oldCommandLine := pflag.CommandLine
		defer func() {
			pflag.CommandLine = oldCommandLine
			pflag.Parse()
		}()

		tests := []struct {
			args []string
			want bool
		}{
			{
				[]string{"minikube", "delete", "--purge"},
				true,
			},
			{
				[]string{"minikube", "delete"},
				false,
			},
			{
				[]string{"minikube", "start", "--purge"},
				false,
			},
			{
				[]string{"minikube"},
				false,
			},
		}

		for _, test := range tests {
			mockArgs(t, test.args)

			got := isDeletePurge()
			if got != test.want {
				t.Errorf("test.args = %q; isDeletePurge() = %t; want %t", test.args, got, test.want)
			}
		}
	})

	// Check if logging with limited args causes a panic
	t.Run("LogCommandStart", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"minikube", "start"}

		oldCommandLine := pflag.CommandLine
		defer func() {
			pflag.CommandLine = oldCommandLine
			pflag.Parse()
		}()
		mockArgs(t, os.Args)
		auditID, err := LogCommandStart()
		if err != nil {
			t.Fatal(err)
		}
		if auditID == "" {
			t.Fatal("audit ID should not be empty")
		}
	})

	t.Run("LogCommandEnd", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"minikube", "start"}
		viper.Set(config.MaxAuditEntries, 3)

		oldCommandLine := pflag.CommandLine
		defer func() {
			pflag.CommandLine = oldCommandLine
			pflag.Parse()
		}()
		mockArgs(t, os.Args)
		auditID, err := LogCommandStart()
		if err != nil {
			t.Fatalf("start failed: %v", err)
		}
		if err := LogCommandEnd(auditID); err != nil {
			t.Fatal(err)
		}

		b, err := exec.Command("wc", "-l", auditOverrideFilename).Output()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "3") {
			t.Errorf("MaxAuditEntries did not work, expected 3 lines in the audit log found %s", string(b))
		}

	})

	t.Run("LogCommandEndNonExistingID", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"minikube", "start"}

		oldCommandLine := pflag.CommandLine
		defer func() {
			pflag.CommandLine = oldCommandLine
			pflag.Parse()
		}()
		mockArgs(t, os.Args)
		if err := LogCommandEnd("non-existing-id"); err == nil {
			t.Fatal("function LogCommandEnd should return an error when a non-existing id is passed in it as an argument")
		}
	})
}

func mockArgs(t *testing.T, args []string) {
	if len(args) == 0 {
		t.Fatalf("cannot pass an empty slice to mockArgs")
	}
	fs := pflag.NewFlagSet(args[0], pflag.ExitOnError)
	fs.Bool("purge", false, "")
	if err := fs.Parse(args[1:]); err != nil {
		t.Fatal(err)
	}
	pflag.CommandLine = fs
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		t.Fatal(err)
	}
}
