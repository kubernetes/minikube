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

// Package name defines structured types for representing image references.
package name

import (
	"strings"
)

const (
	// These have the form: sha256:<hex string>
	// TODO(dekkagaijin): replace with opencontainers/go-digest or docker/distribution's validation.
	digestChars = "sh:0123456789abcdef"
	digestDelim = "@"
)

// Digest stores a digest name in a structured form.
type Digest struct {
	Repository
	digest string
}

// Ensure Digest implements Reference
var _ Reference = (*Digest)(nil)

// Context implements Reference.
func (d Digest) Context() Repository {
	return d.Repository
}

// Identifier implements Reference.
func (d Digest) Identifier() string {
	return d.DigestStr()
}

// DigestStr returns the digest component of the Digest.
func (d Digest) DigestStr() string {
	return d.digest
}

// Name returns the name from which the Digest was derived.
func (d Digest) Name() string {
	return d.Repository.Name() + digestDelim + d.DigestStr()
}

func (d Digest) String() string {
	return d.Name()
}

func checkDigest(name string) error {
	return checkElement("digest", name, digestChars, 7+64, 7+64)
}

// NewDigest returns a new Digest representing the given name, according to the given strictness.
func NewDigest(name string, strict Strictness) (Digest, error) {
	// Split on "@"
	parts := strings.Split(name, digestDelim)
	if len(parts) != 2 {
		return Digest{}, NewErrBadName("a digest must contain exactly one '@' separator (e.g. registry/repository@digest) saw: %s", name)
	}
	base := parts[0]
	digest := parts[1]

	// We don't require a digest, but if we get one check it's valid,
	// even when not being strict.
	// If we are being strict, we want to validate the digest regardless in case
	// it's empty.
	if digest != "" || strict == StrictValidation {
		if err := checkDigest(digest); err != nil {
			return Digest{}, err
		}
	}

	repo, err := NewRepository(base, strict)
	if err != nil {
		return Digest{}, err
	}
	return Digest{repo, digest}, nil
}
