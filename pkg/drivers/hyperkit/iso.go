/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hyperkit

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hooklift/iso9660"
)

// ExtractFile extracts a file from an ISO
func ExtractFile(isoPath, srcPath, destPath string) error {
	iso, err := os.Open(isoPath)
	defer iso.Close()
	if err != nil {
		return err
	}

	r, err := iso9660.NewReader(iso)
	if err != nil {
		return err
	}

	f, err := findFile(r, srcPath)
	if err != nil {
		return err
	}

	dst, err := os.Create(destPath)
	defer dst.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, f.Sys().(io.Reader))
	return err
}

func readFile(isoPath, srcPath string) (string, error) {
	iso, err := os.Open(isoPath)
	defer iso.Close()
	if err != nil {
		return "", err
	}

	r, err := iso9660.NewReader(iso)
	if err != nil {
		return "", err
	}

	f, err := findFile(r, srcPath)
	if err != nil {
		return "", err
	}

	contents, err := ioutil.ReadAll(f.Sys().(io.Reader))
	return string(contents), err
}

func findFile(r *iso9660.Reader, path string) (os.FileInfo, error) {
	// Look through the ISO for a file with a matching path.
	for f, err := r.Next(); err != io.EOF; f, err = r.Next() {
		// For some reason file paths in the ISO sometimes contain a '.' character at the end, so strip that off.
		if strings.TrimSuffix(f.Name(), ".") == path {
			return f, nil
		}
	}
	return nil, fmt.Errorf("unable to find file %s", path)
}
