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

package v1util

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type verifyReader struct {
	inner    io.Reader
	hasher   hash.Hash
	expected v1.Hash
}

// Read implements io.Reader
func (vc *verifyReader) Read(b []byte) (int, error) {
	n, err := vc.inner.Read(b)
	if err == io.EOF {
		got := hex.EncodeToString(vc.hasher.Sum(make([]byte, 0, vc.hasher.Size())))
		if want := vc.expected.Hex; got != want {
			return n, fmt.Errorf("error verifying %s checksum; got %q, want %q",
				vc.expected.Algorithm, got, want)
		}
	}
	return n, err
}

// VerifyReadCloser wraps the given io.ReadCloser to verify that its contents match
// the provided v1.Hash before io.EOF is returned.
func VerifyReadCloser(r io.ReadCloser, h v1.Hash) (io.ReadCloser, error) {
	w, err := v1.Hasher(h.Algorithm)
	if err != nil {
		return nil, err
	}
	r2 := io.TeeReader(r, w)
	return &readAndCloser{
		Reader: &verifyReader{
			inner:    r2,
			hasher:   w,
			expected: h,
		},
		CloseFunc: r.Close,
	}, nil
}
