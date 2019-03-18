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

package tarball

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

type layer struct {
	digest     v1.Hash
	diffID     v1.Hash
	size       int64
	opener     Opener
	compressed bool
}

func (l *layer) Digest() (v1.Hash, error) {
	return l.digest, nil
}

func (l *layer) DiffID() (v1.Hash, error) {
	return l.diffID, nil
}

func (l *layer) Compressed() (io.ReadCloser, error) {
	rc, err := l.opener()
	if err == nil && !l.compressed {
		return v1util.GzipReadCloser(rc)
	}

	return rc, err
}

func (l *layer) Uncompressed() (io.ReadCloser, error) {
	rc, err := l.opener()
	if err == nil && l.compressed {
		return v1util.GunzipReadCloser(rc)
	}

	return rc, err
}

func (l *layer) Size() (int64, error) {
	return l.size, nil
}

// LayerFromFile returns a v1.Layer given a tarball
func LayerFromFile(path string) (v1.Layer, error) {
	opener := func() (io.ReadCloser, error) {
		return os.Open(path)
	}
	return LayerFromOpener(opener)
}

// LayerFromOpener returns a v1.Layer given an Opener function
func LayerFromOpener(opener Opener) (v1.Layer, error) {
	rc, err := opener()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	compressed, err := v1util.IsGzipped(rc)
	if err != nil {
		return nil, err
	}

	var digest v1.Hash
	var size int64
	if digest, size, err = computeDigest(opener, compressed); err != nil {
		return nil, err
	}

	diffID, err := computeDiffID(opener, compressed)
	if err != nil {
		return nil, err
	}

	return &layer{
		digest:     digest,
		diffID:     diffID,
		size:       size,
		compressed: compressed,
		opener:     opener,
	}, nil
}

func computeDigest(opener Opener, compressed bool) (v1.Hash, int64, error) {
	rc, err := opener()
	if err != nil {
		return v1.Hash{}, 0, err
	}
	defer rc.Close()

	if compressed {
		return v1.SHA256(rc)
	}

	reader, err := v1util.GzipReadCloser(ioutil.NopCloser(rc))
	if err != nil {
		return v1.Hash{}, 0, err
	}

	return v1.SHA256(reader)
}

func computeDiffID(opener Opener, compressed bool) (v1.Hash, error) {
	rc, err := opener()
	if err != nil {
		return v1.Hash{}, err
	}
	defer rc.Close()

	if !compressed {
		digest, _, err := v1.SHA256(rc)
		return digest, err
	}

	reader, err := gzip.NewReader(rc)
	if err != nil {
		return v1.Hash{}, err
	}

	diffID, _, err := v1.SHA256(reader)
	return diffID, err
}
