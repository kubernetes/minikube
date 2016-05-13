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

package tarball

import (
	"archive/tar"
	"io"
)

// TarFile is a representation of a file in a tarball. It consists of two parts,
// the Header and the Stream. The Header is a regular tar header, the Stream
// is a byte stream that can be used to read the file's contents
type TarFile struct {
	Header    *tar.Header
	TarStream io.Reader
}

// Name returns the name of the file as reported by the header
func (t *TarFile) Name() string {
	return t.Header.Name
}

// Linkname returns the Linkname of the file as reported by the header
func (t *TarFile) Linkname() string {
	return t.Header.Linkname
}
