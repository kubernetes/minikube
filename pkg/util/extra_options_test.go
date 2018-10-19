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
