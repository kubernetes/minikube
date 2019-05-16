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

package util

import (
	"math"
	"net"
	"reflect"
	"testing"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
)

type aliasedString string

type testConfig struct {
	A string
	B int
	C float32
	D subConfig1
	E *subConfig2
}

type subConfig1 struct {
	F string
	G int
	H float32
	I subConfig3
}

type subConfig2 struct {
	J string
	K int
	L float32
}

type subConfig3 struct {
	M string
	N int
	O float32
	P bool
	Q net.IP
	R utilnet.PortRange
	S []string
	T aliasedString
	U net.IPNet
	V time.Duration
}

func buildConfig() testConfig {
	_, cidr, _ := net.ParseCIDR("12.34.56.78/16")
	return testConfig{
		A: "foo",
		B: 1,
		C: 1.1,
		D: subConfig1{
			F: "bar",
			G: 2,
			H: 2.2,
			I: subConfig3{
				M: "baz",
				N: 3,
				O: 3.3,
				P: false,
				Q: net.ParseIP("12.34.56.78"),
				R: utilnet.PortRange{Base: 2, Size: 4},
				U: *cidr,
				V: 5 * time.Second,
			},
		},
		E: &subConfig2{
			J: "bat",
			K: 4,
			L: 4.4,
		},
	}
}

func TestFindNestedStrings(t *testing.T) {
	a := buildConfig()
	for _, tc := range []struct {
		input  string
		output string
	}{
		{"A", "foo"},
		{"D.F", "bar"},
		{"D.I.M", "baz"},
		{"E.J", "bat"},
	} {
		v, err := findNestedElement(tc.input, &a)
		if err != nil {
			t.Fatalf("Did not expect error. Got: %v", err)
		}
		if v.String() != tc.output {
			t.Fatalf("Expected: %s, got %s", tc.output, v.String())
		}
	}
}

func TestFindNestedInts(t *testing.T) {
	a := buildConfig()

	for _, tc := range []struct {
		input  string
		output int64
	}{
		{"B", 1},
		{"D.G", 2},
		{"D.I.N", 3},
		{"E.K", 4},
	} {
		v, err := findNestedElement(tc.input, &a)
		if err != nil {
			t.Fatalf("Did not expect error. Got: %v", err)
		}
		if v.Int() != tc.output {
			t.Fatalf("Expected: %d, got %d", tc.output, v.Int())
		}
	}
}

func checkFloats(f1, f2 float64) bool {
	return math.Abs(f1-f2) < .00001
}

func TestFindNestedFloats(t *testing.T) {
	a := buildConfig()
	for _, tc := range []struct {
		input  string
		output float64
	}{
		{"C", 1.1},
		{"D.H", 2.2},
		{"D.I.O", 3.3},
		{"E.L", 4.4},
	} {
		v, err := findNestedElement(tc.input, &a)
		if err != nil {
			t.Fatalf("Did not expect error. Got: %v", err)
		}

		// Floating point comparison is tricky.
		if !checkFloats(tc.output, v.Float()) {
			t.Fatalf("Expected: %v, got %v", tc.output, v.Float())
		}
	}
}

func TestSetElement(t *testing.T) {
	for _, tc := range []struct {
		path    string
		newval  string
		checker func(testConfig) bool
	}{
		{"A", "newstring", func(t testConfig) bool { return t.A == "newstring" }},
		{"B", "13", func(t testConfig) bool { return t.B == 13 }},
		{"C", "3.14", func(t testConfig) bool { return checkFloats(float64(t.C), 3.14) }},
		{"D.F", "fizzbuzz", func(t testConfig) bool { return t.D.F == "fizzbuzz" }},
		{"D.G", "4", func(t testConfig) bool { return t.D.G == 4 }},
		{"D.H", "7.3", func(t testConfig) bool { return checkFloats(float64(t.D.H), 7.3) }},
		{"E.J", "otherstring", func(t testConfig) bool { return t.E.J == "otherstring" }},
		{"E.K", "17", func(t testConfig) bool { return t.E.K == 17 }},
		{"E.L", "1.234", func(t testConfig) bool { return checkFloats(float64(t.E.L), 1.234) }},
		{"D.I.P", "true", func(t testConfig) bool { return t.D.I.P == true }},
		{"D.I.P", "false", func(t testConfig) bool { return t.D.I.P == false }},
		{"D.I.Q", "11.22.33.44", func(t testConfig) bool { return t.D.I.Q.Equal(net.ParseIP("11.22.33.44")) }},
		{"D.I.R", "7-11", func(t testConfig) bool { return t.D.I.R.Base == 7 && t.D.I.R.Size == 5 }},
		{"D.I.S", "a,b", func(t testConfig) bool { return reflect.DeepEqual(t.D.I.S, []string{"a", "b"}) }},
		{"D.I.T", "foo", func(t testConfig) bool { return t.D.I.T == "foo" }},
		{"D.I.U", "11.22.0.0/16", func(t testConfig) bool { return t.D.I.U.String() == "11.22.0.0/16" }},
		{"D.I.V", "5s", func(t testConfig) bool { return t.D.I.V == 5*time.Second }},
	} {
		a := buildConfig()
		if err := FindAndSet(tc.path, &a, tc.newval); err != nil {
			t.Fatalf("Error setting value: %v", err)
		}
		if !tc.checker(a) {
			t.Fatalf("Error, values not correct: %v, %s, %s", a, tc.newval, tc.path)
		}

	}
}
