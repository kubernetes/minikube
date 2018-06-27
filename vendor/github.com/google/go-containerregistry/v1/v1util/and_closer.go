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

// readAndCloser implements io.ReadCloser by reading from a particular io.Reader
// and then calling the provided "Close()" method.
type readAndCloser struct {
	io.Reader
	CloseFunc func() error
}

var _ io.ReadCloser = (*readAndCloser)(nil)

// Close implements io.ReadCloser
func (rac *readAndCloser) Close() error {
	return rac.CloseFunc()
}

// writeAndCloser implements io.WriteCloser by reading from a particular io.Writer
// and then calling the provided "Close()" method.
type writeAndCloser struct {
	io.Writer
	CloseFunc func() error
}

var _ io.WriteCloser = (*writeAndCloser)(nil)

// Close implements io.WriteCloser
func (wac *writeAndCloser) Close() error {
	return wac.CloseFunc()
}
