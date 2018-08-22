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

package name

import (
	"errors"
	"fmt"
)

// Reference defines the interface that consumers use when they can
// take either a tag or a digest.
type Reference interface {
	fmt.Stringer

	// Context accesses the Repository context of the reference.
	Context() Repository

	// Identifier accesses the type-specific portion of the reference.
	Identifier() string

	// Name is the fully-qualified reference name.
	Name() string

	// Scope is the scope needed to access this reference.
	Scope(string) string
}

// ParseReference parses the string as a reference, either by tag or digest.
func ParseReference(s string, strict Strictness) (Reference, error) {
	if t, err := NewTag(s, strict); err == nil {
		return t, nil
	}
	if d, err := NewDigest(s, strict); err == nil {
		return d, nil
	}
	// TODO: Combine above errors into something more useful?
	return nil, errors.New("could not parse reference")
}
