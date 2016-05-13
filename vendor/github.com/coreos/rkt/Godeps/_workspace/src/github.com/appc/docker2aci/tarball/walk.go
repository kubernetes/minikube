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
	"fmt"
	"io"
)

// WalkFunc is a func for handling each file (header and byte stream) in a tarball
type WalkFunc func(t *TarFile) error

// Walk walks through the files in the tarball represented by tarstream and
// passes each of them to the WalkFunc provided as an argument
func Walk(tarReader tar.Reader, walkFunc func(t *TarFile) error) error {
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return fmt.Errorf("Error reading tar entry: %v", err)
		}
		if err := walkFunc(&TarFile{Header: hdr, TarStream: &tarReader}); err != nil {
			return err
		}
	}
	return nil
}
