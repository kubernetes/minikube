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

package proxy

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"k8s.io/client-go/rest"
)

func TestIsValidEnv(t *testing.T) {
	var testCases = []struct {
		env  string
		want bool
	}{
		{"", false},
		{"HTTPS-PROXY", false},
		{"NOPROXY", false},
		{"http_proxy", true},
	}
	for _, tc := range testCases {
		t.Run(tc.env, func(t *testing.T) {
			if got := isValidEnv(tc.env); got != tc.want {
				t.Errorf("isValidEnv(\"%v\") got %v; want %v", tc.env, got, tc.want)
			}
		})
	}

}
func TestIsInBlock(t *testing.T) {

	var testCases = []struct {
		ip        string
		block     string
		want      bool
		wanntAErr bool
	}{
		{"", "192.168.0.1/32", false, true},
		{"192.168.0.1", "", false, true},
		{"192.168.0.1", "192.168.0.1", true, false},
		{"192.168.0.1", "192.168.0.1/32", true, false},
		{"192.168.0.2", "192.168.0.1/32", false, true},
		{"192.168.0.1", "192.168.0.1/18", true, false},
		{"abcd", "192.168.0.1/18", false, true},
		{"192.168.0.1", "foo", false, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %s Want: %t WantAErr: %t", tc.ip, tc.block, tc.want, tc.wanntAErr), func(t *testing.T) {
			got, err := isInBlock(tc.ip, tc.block)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != tc.wanntAErr {
				t.Errorf("isInBlock(%v,%v) got error is %v ; want error is %v", tc.ip, tc.block, gotErr, tc.wanntAErr)
			}

			if got != tc.want {
				t.Errorf("isInBlock(%v,%v) got %v; want %v", tc.ip, tc.block, got, tc.want)
			}

		})
	}
}

func TestUpdateEnv(t *testing.T) {
	var testCases = []struct {
		ip      string
		env     string
		wantErr bool
	}{
		{"192.168.0.13", "NO_PROXY", false},
		{"", "NO_PROXY", true},
		{"", "", true},
		{"192.168.0.13", "", true},
		{"192.168.0.13", "NPROXY", true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %s WantAErr: %t", tc.ip, tc.env, tc.wantErr), func(t *testing.T) {
			origVal := os.Getenv(tc.env)
			gotErr := false
			err := updateEnv(tc.ip, tc.env)
			if err != nil {
				gotErr = true
			}
			if gotErr != tc.wantErr {
				t.Errorf("updateEnv(%v,%v) got error is %v ; want error is %v", tc.ip, tc.env, gotErr, tc.wantErr)
			}
			err = os.Setenv(tc.env, origVal)
			if err != nil && tc.env != "" {
				t.Errorf("Error reverting the env var (%s) to its original value (%s)", tc.env, origVal)
			}

		})
	}

}

func TestCheckEnv(t *testing.T) {
	var testCases = []struct {
		ip           string
		envName      string
		want         bool
		mockEnvValue string
	}{
		{"", "NO_PROXY", false, ""},
		{"192.168.0.13", "NO_PROXY", false, ""},
		{"192.168.0.13", "NO_PROXY", false, ","},
		{"192.168.0.13", "NO_PROXY", true, "192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", false, "192.168.0.14"},
		{"192.168.0.13", "NO_PROXY", true, ",192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", true, "10.10.0.13,192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", true, "192.168.0.13/22"},
		{"192.168.0.13", "NO_PROXY", true, "10.10.0.13,192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", false, "10.10.0.13/22"},
		{"10.10.10.4", "NO_PROXY", true, "172.168.0.0/30,10.10.10.0/24"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %s", tc.ip, tc.envName), func(t *testing.T) {
			t.Setenv(tc.envName, tc.mockEnvValue)
			if got := checkEnv(tc.ip, tc.envName); got != tc.want {
				t.Errorf("CheckEnv(%v,%v) got  %v ; want is %v", tc.ip, tc.envName, got, tc.want)
			}
		})
	}

}

func TestIsIPExcluded(t *testing.T) {
	var testCases = []struct {
		ip, env  string
		excluded bool
	}{
		{"1.2.3.4", "7.7.7.7", false},
		{"1.2.3.4", "1.2.3.4", true},
		{"1.2.3.4", "", false},
		{"foo", "", false},
		{"foo", "bar", false},
		{"foo", "1.2.3.4", false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("exclude %s NO_PROXY(%v)", tc.ip, tc.env), func(t *testing.T) {
			t.Setenv("NO_PROXY", tc.env)
			if excluded := IsIPExcluded(tc.ip); excluded != tc.excluded {
				t.Fatalf("IsIPExcluded(%v) should return %v. NO_PROXY=%v", tc.ip, tc.excluded, tc.env)
			}
		})
	}
}

func TestExcludeIP(t *testing.T) {
	var testCases = []struct {
		ip, env  string
		wantAErr bool
	}{
		{"1.2.3.4", "", false},
		{"", "", true},
		{"7.7.7.7", "7.7.7.7", false},
		{"7.7.7.7", "1.2.3.4", false},
		{"foo", "", true},
		{"foo", "1.2.3.4", true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("exclude %s NO_PROXY(%s)", tc.ip, tc.env), func(t *testing.T) {
			t.Setenv("NO_PROXY", tc.env)
			err := ExcludeIP(tc.ip)
			if err != nil && !tc.wantAErr {
				t.Errorf("ExcludeIP(%v) returned unexpected error %v", tc.ip, err)
			}
			if err == nil && tc.wantAErr {
				t.Errorf("ExcludeIP(%v) should return error but error is %v", tc.ip, err)
			}
		})
	}
}

func TestUpdateTransport(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		rcBefore := rest.Config{}
		rcAfter := UpdateTransport(&rcBefore)
		if rcAfter.WrapTransport == nil {
			t.Errorf("UpdateTransport(%v) = %v, WrapTransport was unexpectedly nil", rcBefore, rcAfter)
		}
	})
	t.Run("existing", func(t *testing.T) {
		// rest config initialized with WrapTransport function
		rc := rest.Config{WrapTransport: func(http.RoundTripper) http.RoundTripper {
			rt := &http.Transport{}
			rt.RegisterProtocol("file", http.NewFileTransport(http.Dir("/tmp")))
			return rt
		}}

		transport := &http.Transport{}
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

		c := UpdateTransport(&rc)
		rt := c.WrapTransport(nil)

		if rt == rc.WrapTransport(transport) {
			t.Fatalf("Expected to reuse existing RoundTripper(%v) but found %v", rt, transport)
		}

	})
	t.Run("nil", func(t *testing.T) {
		rc := rest.Config{}
		c := UpdateTransport(&rc)
		rt := c.WrapTransport(nil)
		if rt != nil {
			t.Fatalf("Expected RoundTripper nil for invocation WrapTransport(nil)")
		}
	})
}
func TestMaskProxyPassword(t *testing.T) {
	type dockerOptTest struct {
		input  string
		output string
	}
	var tests = []dockerOptTest{
		{
			input:  "cats",
			output: "cats",
		},
		{
			input:  "myDockerOption=value",
			output: "myDockerOption=value",
		},
		{
			input:  "http://minikube.sigs.k8s.io",
			output: "http://minikube.sigs.k8s.io",
		},
		{
			input:  "http://jdoe@minikube.sigs.k8s.io:8080",
			output: "http://jdoe@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "https://mary:iam$Fake!password@minikube.sigs.k8s.io:8080",
			output: "https://mary:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http://jdoe:%n0tRe@al:Password!@minikube.sigs.k8s.io:8080",
			output: "http://jdoe:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http://jo@han:n0tRe@al:&Password!@minikube.sigs.k8s.io:8080",
			output: "http://jo@han:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http://k@r3n!:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "http://k@r3n!:*****@minikube.sigs.k8s.io",
		},
		{
			input:  "https://fr@ank5t3in:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "https://fr@ank5t3in:*****@minikube.sigs.k8s.io",
		}, {
			input:  "http://aaa:bbb@internet-app.corp.abcd:8080",
			output: "http://aaa:*****@internet-app.corp.abcd:8080",
		},
		{
			input:  "https://aaa:bbb@internet-app.corp.abcd:8080",
			output: "https://aaa:*****@internet-app.corp.abcd:8080",
		},
	}
	for _, test := range tests {
		got := MaskProxyPassword(test.input)
		if got != test.output {
			t.Errorf("MaskProxyPassword(\"%v\"): got %v, expected %v", test.input, got, test.output)
		}
	}
}

func TestMaskProxyPasswordWithKey(t *testing.T) {
	type dockerOptTest struct {
		input  string
		output string
	}
	var tests = []dockerOptTest{
		{
			input:  "cats",
			output: "cats",
		},
		{
			input:  "myDockerOption=value",
			output: "myDockerOption=value",
		},
		{
			input:  "http_proxy=http://minikube.sigs.k8s.io",
			output: "HTTP_PROXY=http://minikube.sigs.k8s.io",
		},
		{
			input:  "https_proxy=http://jdoe@minikube.sigs.k8s.io:8080",
			output: "HTTPS_PROXY=http://jdoe@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "https_proxy=https://mary:iam$Fake!password@minikube.sigs.k8s.io:8080",
			output: "HTTPS_PROXY=https://mary:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://jdoe:%n0tRe@al:Password!@minikube.sigs.k8s.io:8080",
			output: "HTTP_PROXY=http://jdoe:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://jo@han:n0tRe@al:&Password!@minikube.sigs.k8s.io:8080",
			output: "HTTP_PROXY=http://jo@han:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://k@r3n!:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "HTTP_PROXY=http://k@r3n!:*****@minikube.sigs.k8s.io",
		},
		{
			input:  "https_proxy=https://fr@ank5t3in:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "HTTPS_PROXY=https://fr@ank5t3in:*****@minikube.sigs.k8s.io",
		},
	}
	for _, test := range tests {
		got := MaskProxyPasswordWithKey(test.input)
		if got != test.output {
			t.Errorf("MaskProxyPasswordWithKey(\"%v\"): got %v, expected %v", test.input, got, test.output)
		}
	}
}
