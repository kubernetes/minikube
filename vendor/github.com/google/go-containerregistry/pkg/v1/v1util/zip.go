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
	"bytes"
	"compress/gzip"
	"io"
)

var gzipMagicHeader = []byte{'\x1f', '\x8b'}

// GzipReadCloser reads uncompressed input data from the io.ReadCloser and
// returns an io.ReadCloser from which compressed data may be read.
// This uses gzip.BestSpeed for the compression level.
func GzipReadCloser(r io.ReadCloser) (io.ReadCloser, error) {
	return GzipReadCloserLevel(r, gzip.BestSpeed)
}

// GzipReadCloserLevel reads uncompressed input data from the io.ReadCloser and
// returns an io.ReadCloser from which compressed data may be read.
// Refer to compress/gzip for the level:
// https://golang.org/pkg/compress/gzip/#pkg-constants
func GzipReadCloserLevel(r io.ReadCloser, level int) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer r.Close()

		gw, _ := gzip.NewWriterLevel(pw, level)
		defer gw.Close()

		_, err := io.Copy(gw, r)
		if err != nil {
			pr.CloseWithError(err)
		}
	}()

	return pr, nil
}

// GunzipReadCloser reads compressed input data from the io.ReadCloser and
// returns an io.ReadCloser from which uncompessed data may be read.
func GunzipReadCloser(r io.ReadCloser) (io.ReadCloser, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &readAndCloser{
		Reader: gr,
		CloseFunc: func() error {
			if err := gr.Close(); err != nil {
				return err
			}
			return r.Close()
		},
	}, nil
}

// IsGzipped detects whether the input stream is compressed.
func IsGzipped(r io.Reader) (bool, error) {
	magicHeader := make([]byte, 2)
	n, err := r.Read(magicHeader)
	if n == 0 && err == io.EOF {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return bytes.Equal(magicHeader, gzipMagicHeader), nil
}
