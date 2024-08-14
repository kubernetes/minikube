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

package out

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/spf13/pflag"
	"golang.org/x/text/language"

	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/minikube/translate"
)

func TestStep(t *testing.T) {
	// Set the system locale to Arabic and define a dummy translation file.
	translate.SetPreferredLanguage(language.Arabic)

	translate.Translations = map[string]interface{}{
		"Installing Kubernetes version {{.version}} ...": "... {{.version}} تثبيت Kubernetes الإصدار",
	}

	testCases := []struct {
		style     style.Enum
		message   string
		params    V
		want      string
		wantASCII string
	}{
		{style.Happy, "Happy", nil, "😄  Happy\n", "* Happy\n"},
		{style.Option, "Option", nil, "    ▪ Option\n", "  - Option\n"},
		{style.Warning, "Warning", nil, "❗  Warning\n", "! Warning\n"},
		{style.Fatal, "Fatal: {{.error}}", V{"error": "ugh"}, "💣  Fatal: ugh\n", "X Fatal: ugh\n"},
		{style.Issue, "http://i/{{.number}}", V{"number": 10000}, "    ▪ http://i/10000\n", "  - http://i/10000\n"},
		{style.Usage, "raw: {{.one}} {{.two}}", V{"one": "'%'", "two": "%d"}, "💡  raw: '%' %d\n", "* raw: '%' %d\n"},
		{style.Running, "Installing Kubernetes version {{.version}} ...", V{"version": "v1.13"}, "🏃  ... v1.13 تثبيت Kubernetes الإصدار\n", "* ... v1.13 تثبيت Kubernetes الإصدار\n"},
	}
	for _, tc := range testCases {
		for _, override := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s-override-%v", tc.message, override), func(t *testing.T) {
				// Set MINIKUBE_IN_STYLE=<override>
				t.Setenv(OverrideEnv, strconv.FormatBool(override))
				f := tests.NewFakeFile()
				SetOutFile(f)
				Step(tc.style, tc.message, tc.params)
				got := f.String()
				want := tc.wantASCII
				if override {
					want = tc.want
				}
				if got != want {
					t.Errorf("Step() = %q (%d runes), want %q (%d runes)", got, len(got), want, len(want))
				}
			})
		}
	}
}

func TestString(t *testing.T) {
	t.Setenv(OverrideEnv, "")

	testCases := []struct {
		format string
		arg    interface{}
		want   string
	}{
		{format: "xyz123", want: "xyz123"},
		{format: "Installing Kubernetes version %s ...", arg: "v1.13", want: "Installing Kubernetes version v1.13 ..."},
		{format: "Parameter encoding: %s", arg: "%s%%%d", want: "Parameter encoding: %s%%%d"},
	}
	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			f := tests.NewFakeFile()
			SetOutFile(f)
			ErrLn("unrelated message")
			if tc.arg == nil {
				String(tc.format)
			} else {
				Stringf(tc.format, tc.arg)
			}
			got := f.String()
			if got != tc.want {
				t.Errorf("String(%s, %s) = %q, want %q", tc.format, tc.arg, got, tc.want)
			}
		})
	}
}

func TestErr(t *testing.T) {
	t.Setenv(OverrideEnv, "0")
	f := tests.NewFakeFile()
	SetErrFile(f)
	Err("xyz123\n")
	Ln("unrelated message")
	got := f.String()
	want := "xyz123\n"

	if got != want {
		t.Errorf("Err() = %q, want %q", got, want)
	}
}

func TestErrf(t *testing.T) {
	t.Setenv(OverrideEnv, "0")
	f := tests.NewFakeFile()
	SetErrFile(f)
	Errf("xyz123 %s\n", "%s%%%d")
	Ln("unrelated message")
	got := f.String()
	want := "xyz123 %s%%%d\n"

	if got != want {
		t.Errorf("Errf() = %q, want %q", got, want)
	}
}

func createLogFile() (string, error) {
	td := os.TempDir()
	name := filepath.Join(td, "minikube_test_test_test.log")
	f, err := os.Create(name)
	if err != nil {
		return "", fmt.Errorf("failed to create log file: %v", err)
	}

	return f.Name(), nil
}

func TestLatestLogFilePath(t *testing.T) {
	want, err := createLogFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(want)

	got, err := latestLogFilePath()
	if err != nil {
		t.Errorf("latestLogFilePath() failed with error = %v", err)
	}
	if got != want {
		t.Errorf("latestLogFilePath() = %q; wanted %q", got, want)
	}
}

func TestCommand(t *testing.T) {
	testCases := []struct {
		args        []string
		want        string
		shouldError bool
	}{
		{
			[]string{"minikube", "start"},
			"start",
			false,
		},
		{
			[]string{"minikube", "--profile", "profile1", "start"},
			"start",
			false,
		},
		{
			[]string{"minikube"},
			"",
			true,
		},
	}

	pflag.String("profile", "", "")

	for _, tt := range testCases {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = tt.args
		pflag.Parse()
		got, err := command()
		if err == nil && tt.shouldError {
			t.Errorf("os.Args = %s; command() did not fail but was expected to", tt.args)
		}
		if err != nil && !tt.shouldError {
			t.Errorf("os.Args = %s; command() failed with error = %v", tt.args, err)
		}
		if got != tt.want {
			t.Errorf("os.Args = %s; command() = %q; wanted %q", tt.args, got, tt.want)
		}
	}
}

func TestDisplayGitHubIssueMessage(t *testing.T) {
	testCases := []struct {
		args                 []string
		shouldContainMessage bool
	}{
		{
			[]string{"minikube", "start"},
			false,
		},
		{
			[]string{"minikube", "delete"},
			true,
		},
	}

	msg := "Please also attach the following file to the GitHub issue:"

	for _, tt := range testCases {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = tt.args
		pflag.Parse()
		f := tests.NewFakeFile()
		SetErrFile(f)
		displayGitHubIssueMessage()
		output := f.String()
		if strings.Contains(output, msg) && !tt.shouldContainMessage {
			t.Errorf("os.Args = %s; displayGitHubIssueMessage() output = %q; did not expect it to contain = %q", tt.args, output, msg)
		}
		if !strings.Contains(output, msg) && tt.shouldContainMessage {
			t.Errorf("os.Args = %s; displayGitHubIssueMessage() output = %q; expected to contain = %q", tt.args, output, msg)
		}
	}
}

func TestBoxed(t *testing.T) {
	f := tests.NewFakeFile()
	SetOutFile(f)
	Boxed(`Running with {{.driver}} driver and port {{.port}}`, V{"driver": "docker", "port": 8000})
	got := f.String()
	want :=
		`╭────────────────────────────────────────────────╮
│                                                │
│    Running with docker driver and port 8000    │
│                                                │
╰────────────────────────────────────────────────╯
`
	if got != want {
		t.Errorf("Boxed() = %q, want %q", got, want)
	}
}

func TestBoxedErr(t *testing.T) {
	f := tests.NewFakeFile()
	SetErrFile(f)
	BoxedErr(`Running with {{.driver}} driver and port {{.port}}`, V{"driver": "docker", "port": 8000})
	got := f.String()
	want :=
		`╭────────────────────────────────────────────────╮
│                                                │
│    Running with docker driver and port 8000    │
│                                                │
╰────────────────────────────────────────────────╯
`
	if got != want {
		t.Errorf("Boxed() = %q, want %q", got, want)
	}
}

func TestBoxedWithConfig(t *testing.T) {
	testCases := []struct {
		config box.Config
		st     style.Enum
		title  string
		format string
		args   []V
		want   string
	}{
		{
			box.Config{Px: 2, Py: 2},
			style.None,
			"",
			"Boxed content",
			nil,
			`┌─────────────────┐
│                 │
│                 │
│  Boxed content  │
│                 │
│                 │
└─────────────────┘
`,
		},
		{
			box.Config{Px: 0, Py: 0},
			style.None,
			"",
			"Boxed content with 0 padding",
			nil,
			`┌────────────────────────────┐
│Boxed content with 0 padding│
└────────────────────────────┘
`,
		},
		{
			box.Config{Px: 1, Py: 1, TitlePos: "Inside"},
			style.None,
			"Hello World",
			"Boxed content with title inside",
			nil,
			`┌─────────────────────────────────┐
│                                 │
│           Hello World           │
│                                 │
│ Boxed content with title inside │
│                                 │
└─────────────────────────────────┘
`,
		},
		{
			box.Config{Px: 1, Py: 1, TitlePos: "Top"},
			style.None,
			"Hello World",
			"Boxed content with title inside",
			nil,
			`┌ Hello World ────────────────────┐
│                                 │
│ Boxed content with title inside │
│                                 │
└─────────────────────────────────┘
`,
		},
		{
			box.Config{Px: 1, Py: 1, TitlePos: "Top"},
			style.Tip,
			"Hello World",
			"Boxed content with title inside",
			nil,
			`┌ * Hello World ──────────────────┐
│                                 │
│ Boxed content with title inside │
│                                 │
└─────────────────────────────────┘
`,
		},
		{
			box.Config{Px: 1, Py: 1, TitlePos: "Top"},
			style.Tip,
			// This case is to make sure newlines (\n) are removed before printing
			// Otherwise box-cli-maker panices:
			// https://github.com/Delta456/box-cli-maker/blob/7b5a1ad8a016ce181e7d8b05e24b54ff60b4b38a/box.go#L69-L71
			"Hello \nWorld",
			"Boxed content with title inside",
			nil,
			`┌ * Hello World ──────────────────┐
│                                 │
│ Boxed content with title inside │
│                                 │
└─────────────────────────────────┘
`,
		},
	}

	for _, tc := range testCases {
		f := tests.NewFakeFile()
		SetOutFile(f)
		BoxedWithConfig(tc.config, tc.st, tc.title, tc.format, tc.args...)
		got := f.String()
		if tc.want != got {
			t.Errorf("Expecting BoxedWithConfig(%v, %v, %s, %s, %s) = \n%s, want \n%s", tc.config, tc.st, tc.title, tc.format, tc.args, got, tc.want)
		}
	}
}
