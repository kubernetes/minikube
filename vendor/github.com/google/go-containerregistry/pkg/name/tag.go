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
	"strings"
)

const (
	defaultTag = "latest"
	// TODO(dekkagaijin): use the docker/distribution regexes for validation.
	tagChars = "abcdefghijklmnopqrstuvwxyz0123456789_-.ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	tagDelim = ":"
)

// Tag stores a docker tag name in a structured form.
type Tag struct {
	Repository
	tag string
}

// Ensure Tag implements Reference
var _ Reference = (*Tag)(nil)

// Context implements Reference.
func (t Tag) Context() Repository {
	return t.Repository
}

// Identifier implements Reference.
func (t Tag) Identifier() string {
	return t.TagStr()
}

// TagStr returns the tag component of the Tag.
func (t Tag) TagStr() string {
	if t.tag != "" {
		return t.tag
	}
	return defaultTag
}

// Name returns the name from which the Tag was derived.
func (t Tag) Name() string {
	return t.Repository.Name() + tagDelim + t.TagStr()
}

func (t Tag) String() string {
	return t.Name()
}

// Scope returns the scope required to perform the given action on the tag.
func (t Tag) Scope(action string) string {
	return t.Repository.Scope(action)
}

func checkTag(name string) error {
	return checkElement("tag", name, tagChars, 1, 127)
}

// NewTag returns a new Tag representing the given name, according to the given strictness.
func NewTag(name string, strict Strictness) (Tag, error) {
	base := name
	tag := ""

	// Split on ":"
	parts := strings.Split(name, tagDelim)
	// Verify that we aren't confusing a tag for a hostname w/ port for the purposes of weak validation.
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
		base = strings.Join(parts[:len(parts)-1], tagDelim)
		tag = parts[len(parts)-1]
	}

	// We don't require a tag, but if we get one check it's valid,
	// even when not being strict.
	// If we are being strict, we want to validate the tag regardless in case
	// it's empty.
	if tag != "" || strict == StrictValidation {
		if err := checkTag(tag); err != nil {
			return Tag{}, err
		}
	}

	repo, err := NewRepository(base, strict)
	if err != nil {
		return Tag{}, err
	}
	return Tag{repo, tag}, nil
}
