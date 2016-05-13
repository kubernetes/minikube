// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/appc/spec/pkg/acirenderer"
)

var debugEnabled bool

func Quote(l []string) []string {
	var quoted []string

	for _, s := range l {
		quoted = append(quoted, fmt.Sprintf("%q", s))
	}

	return quoted
}

func ReverseImages(s acirenderer.Images) acirenderer.Images {
	var o acirenderer.Images
	for i := len(s) - 1; i >= 0; i-- {
		o = append(o, s[i])
	}

	return o
}

func In(list []string, el string) bool {
	return IndexOf(list, el) != -1
}

func IndexOf(list []string, el string) int {
	for i, x := range list {
		if el == x {
			return i
		}
	}
	return -1
}

func printTo(w io.Writer, i ...interface{}) {
	s := fmt.Sprint(i...)
	fmt.Fprintln(w, strings.TrimSuffix(s, "\n"))
}

func Warn(i ...interface{}) {
	printTo(os.Stderr, i...)
}

func Info(i ...interface{}) {
	printTo(os.Stderr, i...)
}

func Debug(i ...interface{}) {
	if debugEnabled {
		printTo(os.Stderr, i...)
	}
}

func InitDebug() {
	debugEnabled = true
}
