// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/name"
)

// Detect more complex forms of local references.
var reLocal = regexp.MustCompile(`.*\.local(?:host)?(?::\d{1,5})?$`)

// Detect the loopback IP (127.0.0.1)
var reLoopback = regexp.MustCompile(regexp.QuoteMeta("127.0.0.1"))

// Scheme returns https scheme for all the endpoints except localhost.
func Scheme(reg name.Registry) string {
	if strings.HasPrefix(reg.Name(), "localhost:") {
		return "http"
	}
	if reLocal.MatchString(reg.Name()) {
		return "http"
	}
	if reLoopback.MatchString(reg.Name()) {
		return "http"
	}
	return "https"
}
