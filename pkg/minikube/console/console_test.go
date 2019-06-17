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

package console

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/minikube/translate"
)

func TestOutStyle(t *testing.T) {

	var testCases = []struct {
		style     StyleEnum
		message   string
		params    []interface{}
		want      string
		wantASCII string
	}{
		{Happy, "Happy", nil, "😄  Happy\n", "* Happy\n"},
		{Option, "Option", nil, "    ▪ Option\n", "  - Option\n"},
		{WarningType, "Warning", nil, "⚠️  Warning\n", "! Warning\n"},
		{FatalType, "Fatal: %v", []interface{}{"ugh"}, "💣  Fatal: ugh\n", "X Fatal: ugh\n"},
		{WaitingPods, "wait", nil, "⌛  wait", "* wait"},
		{Issue, "http://i/%d", []interface{}{10000}, "    ▪ http://i/10000\n", "  - http://i/10000\n"},
		{Usage, "raw: %s %s", []interface{}{"'%'", "%d"}, "💡  raw: '%' %d\n", "* raw: '%' %d\n"},
	}
	for _, tc := range testCases {
		for _, override := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s-override-%v", tc.message, override), func(t *testing.T) {
				// Set MINIKUBE_IN_STYLE=<override>
				os.Setenv(OverrideEnv, strconv.FormatBool(override))
				f := tests.NewFakeFile()
				SetOutFile(f)
				OutStyle(tc.style, tc.message, tc.params...)
				got := f.String()
				want := tc.wantASCII
				if override {
					want = tc.want
				}
				if got != want {
					t.Errorf("OutStyle() = %q (%d runes), want %q (%d runes)", got, len(got), want, len(want))
				}
			})
		}
	}
}

func TestOut(t *testing.T) {
	os.Setenv(OverrideEnv, "")
	// An example translation just to assert that this code path is executed.
	err := message.SetString(language.Arabic, "Installing Kubernetes version %s ...", "... %s تثبيت Kubernetes الإصدار")
	if err != nil {
		t.Fatalf("setstring: %v", err)
	}

	var testCases = []struct {
		format string
		lang   string
		arg    interface{}
		want   string
	}{
		{format: "xyz123", want: "xyz123"},
		{format: "Installing Kubernetes version %s ...", lang: "ar", arg: "v1.13", want: "... v1.13 تثبيت Kubernetes الإصدار"},
		{format: "Installing Kubernetes version %s ...", lang: "en-us", arg: "v1.13", want: "Installing Kubernetes version v1.13 ..."},
		{format: "Parameter encoding: %s", arg: "%s%%%d", want: "Parameter encoding: %s%%%d"},
	}
	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			translate.SetPreferredLanguage(tc.lang)
			f := tests.NewFakeFile()
			SetOutFile(f)
			ErrLn("unrelated message")
			Out(tc.format, tc.arg)
			got := f.String()
			if got != tc.want {
				t.Errorf("Out(%s, %s) = %q, want %q", tc.format, tc.arg, got, tc.want)
			}
		})
	}
}

func TestErr(t *testing.T) {
	os.Setenv(OverrideEnv, "0")
	f := tests.NewFakeFile()
	SetErrFile(f)
	Err("xyz123 %s\n", "%s%%%d")
	OutLn("unrelated message")
	got := f.String()
	want := "xyz123 %s%%%d\n"

	if got != want {
		t.Errorf("Err() = %q, want %q", got, want)
	}
}

func TestErrStyle(t *testing.T) {
	os.Setenv(OverrideEnv, "1")
	f := tests.NewFakeFile()
	SetErrFile(f)
	ErrStyle(FatalType, "error: %s", "%s%%%d")
	got := f.String()
	want := "💣  error: %s%%%d\n"
	if got != want {
		t.Errorf("ErrStyle() = %q, want %q", got, want)
	}
}

func TestSetPreferredLanguage(t *testing.T) {
	os.Setenv(OverrideEnv, "0")
	var tests = []struct {
		input string
		want  language.Tag
	}{
		{"", language.AmericanEnglish},
		{"C", language.AmericanEnglish},
		{"zh", language.Chinese},
		{"fr_FR.utf8", language.French},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			// Set something so that we can assert change.
			translate.SetPreferredLanguage("is")
			if err := translate.SetPreferredLanguage(tc.input); err != nil {
				t.Errorf("unexpected error: %q", err)
			}

			want, _ := tc.want.Base()
			got, _ := translate.GetPreferredLanguage().Base()
			if got != want {
				t.Errorf("SetPreferredLanguage(%s) = %q, want %q", tc.input, got, want)
			}
		})
	}

}
