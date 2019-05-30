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
	"flag"
	"reflect"
	"testing"
)

func TestInvalidFlags(t *testing.T) {
	for _, tc := range [][]string{
		{"-e", "foo"},
		{"-e", "foo.bar"},
		{"-e", "foo.bar.baz"},
		{"-e", "foo=bar"},
		{"-e", "foo=bar.baz"},
		// Multiple flags
		{"-e", "foo", "-e", "foo"},
		{"-e", "foo", "-e", "foo", "-e", "foo"},
		{"-e", "foo", "-e", "foo.bar=baz"},
		{"-e", "foo", "-e", "foo.bar=baz"},
	} {
		var flags flag.FlagSet
		flags.Init("test", flag.ContinueOnError)

		var e ExtraOptionSlice
		flags.Var(&e, "e", "usage")
		if err := flags.Parse(tc); err == nil {
			t.Errorf("Expected error, got nil: %s", tc)
		}
	}
}

func TestValidFlags(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		values ExtraOptionSlice
	}{
		{
			[]string{"-e", "foo.bar=baz"},
			ExtraOptionSlice{ExtraOption{Component: "foo", Key: "bar", Value: "baz"}},
		},
		{
			[]string{"-e", "foo.bar.baz=bat"},
			ExtraOptionSlice{ExtraOption{Component: "foo", Key: "bar.baz", Value: "bat"}},
		},
		{
			[]string{"-e", "foo.bar=baz", "-e", "foo.bar.baz=bat"},
			ExtraOptionSlice{ExtraOption{Component: "foo", Key: "bar", Value: "baz"}, ExtraOption{Component: "foo", Key: "bar.baz", Value: "bat"}},
		},
	} {
		var flags flag.FlagSet
		flags.Init("test", flag.ContinueOnError)

		var e ExtraOptionSlice
		flags.Var(&e, "e", "usage")
		if err := flags.Parse(tc.args); err != nil {
			t.Errorf("Unexpected error: %v for %s.", err, tc)
		}

		if !reflect.DeepEqual(e, tc.values) {
			t.Errorf("Wrong parsed value. Expected %s, got %s", tc.values, e)
		}
	}
}

func TestGet(t *testing.T) {
	extraOptions := ExtraOptionSlice{
		ExtraOption{Component: "c1", Key: "bar", Value: "c1-bar"},
		ExtraOption{Component: "c1", Key: "bar-baz", Value: "c1-bar-baz"},
		ExtraOption{Component: "c2", Key: "bar", Value: "c2-bar"},
		ExtraOption{Component: "c3", Key: "bar", Value: "c3-bar"},
	}

	for _, tc := range []struct {
		searchKey       string
		searchComponent []string
		expRes          string
		values          ExtraOptionSlice
	}{
		{"nonexistent", nil, "", extraOptions},
		{"nonexistent", []string{"c1"}, "", extraOptions},
		{"bar", []string{"c2"}, "c2-bar", extraOptions},
		{"bar", []string{"c2", "c3"}, "c2-bar", extraOptions},
		{"bar", nil, "c1-bar", extraOptions},
	} {
		if res := tc.values.Get(tc.searchKey, tc.searchComponent...); res != tc.expRes {
			t.Errorf("Unexpected value. Expected %s, got %s", tc.expRes, res)
		}
	}
}

func TestAsMap(t *testing.T) {
	extraOptions := ExtraOptionSlice{
		ExtraOption{Component: "c1", Key: "bar", Value: "c1-bar"},
		ExtraOption{Component: "c1", Key: "bar-baz", Value: "c1-bar-baz"},
		ExtraOption{Component: "c2", Key: "bar", Value: "c2-bar"},
		ExtraOption{Component: "c3", Key: "bar", Value: "c3-bar"},
	}

	expectedRes := ComponentExtraOptionMap{
		"c1": {
			"bar":     "c1-bar",
			"bar-baz": "c1-bar-baz",
		},
		"c2": {
			"bar": "c2-bar",
		},
		"c3": {
			"bar": "c3-bar",
		},
	}

	res := extraOptions.AsMap()

	if !reflect.DeepEqual(expectedRes, res) {
		t.Errorf("Unexpected value. Expected %s, got %s", expectedRes, res)
	}
}
