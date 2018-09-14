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
	"io"
)

func nop() error {
	return nil
}

// NopWriteCloser wraps the io.Writer as an io.WriteCloser with a Close() method that does nothing.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &writeAndCloser{
		Writer:    w,
		CloseFunc: nop,
	}
}

// NopReadCloser wraps the io.Reader as an io.ReadCloser with a Close() method that does nothing.
// This is technically redundant with ioutil.NopCloser, but provided for symmetry and clarity.
func NopReadCloser(r io.Reader) io.ReadCloser {
	return &readAndCloser{
		Reader:    r,
		CloseFunc: nop,
	}
}
