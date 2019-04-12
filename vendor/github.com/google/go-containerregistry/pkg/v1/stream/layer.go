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

package stream

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"sync"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

var (
	// ErrNotComputed is returned when the requested value is not yet
	// computed because the stream has not been consumed yet.
	ErrNotComputed = errors.New("value not computed until stream is consumed")

	// ErrConsumed is returned by Compressed when the underlying stream has
	// already been consumed and closed.
	ErrConsumed = errors.New("stream was already consumed")
)

// Layer is a streaming implementation of v1.Layer.
type Layer struct {
	blob     io.ReadCloser
	consumed bool

	mu             sync.Mutex
	digest, diffID *v1.Hash
	size           int64
}

var _ v1.Layer = (*Layer)(nil)

// NewLayer creates a Layer from an io.ReadCloser.
func NewLayer(rc io.ReadCloser) *Layer { return &Layer{blob: rc} }

// Digest implements v1.Layer.
func (l *Layer) Digest() (v1.Hash, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.digest == nil {
		return v1.Hash{}, ErrNotComputed
	}
	return *l.digest, nil
}

// DiffID implements v1.Layer.
func (l *Layer) DiffID() (v1.Hash, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.diffID == nil {
		return v1.Hash{}, ErrNotComputed
	}
	return *l.diffID, nil
}

// Size implements v1.Layer.
func (l *Layer) Size() (int64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.size == 0 {
		return 0, ErrNotComputed
	}
	return l.size, nil
}

// Uncompressed implements v1.Layer.
func (l *Layer) Uncompressed() (io.ReadCloser, error) {
	return nil, errors.New("NYI: stream.Layer.Uncompressed is not implemented")
}

// Compressed implements v1.Layer.
func (l *Layer) Compressed() (io.ReadCloser, error) {
	if l.consumed {
		return nil, ErrConsumed
	}
	return newCompressedReader(l)
}

type compressedReader struct {
	closer io.Closer // original blob's Closer.

	h, zh hash.Hash // collects digests of compressed and uncompressed stream.
	pr    io.Reader
	count *countWriter

	l *Layer // stream.Layer to update upon Close.
}

func newCompressedReader(l *Layer) (*compressedReader, error) {
	h := sha256.New()
	zh := sha256.New()
	count := &countWriter{}

	// gzip.Writer writes to the output stream via pipe, a hasher to
	// capture compressed digest, and a countWriter to capture compressed
	// size.
	pr, pw := io.Pipe()
	zw, err := gzip.NewWriterLevel(io.MultiWriter(pw, zh, count), gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	cr := &compressedReader{
		closer: newMultiCloser(zw, l.blob),
		pr:     pr,
		h:      h,
		zh:     zh,
		count:  count,
		l:      l,
	}
	go func() {
		if _, err := io.Copy(io.MultiWriter(h, zw), l.blob); err != nil {
			pw.CloseWithError(err)
			return
		}
		// Now close the compressed reader, to flush the gzip stream
		// and calculate digest/diffID/size. This will cause pr to
		// return EOF which will cause readers of the Compressed stream
		// to finish reading.
		pw.CloseWithError(cr.Close())
	}()

	return cr, nil
}

func (cr *compressedReader) Read(b []byte) (int, error) { return cr.pr.Read(b) }

func (cr *compressedReader) Close() error {
	cr.l.mu.Lock()
	defer cr.l.mu.Unlock()

	// Close the inner ReadCloser.
	if err := cr.closer.Close(); err != nil {
		return err
	}

	diffID, err := v1.NewHash("sha256:" + hex.EncodeToString(cr.h.Sum(nil)))
	if err != nil {
		return err
	}
	cr.l.diffID = &diffID

	digest, err := v1.NewHash("sha256:" + hex.EncodeToString(cr.zh.Sum(nil)))
	if err != nil {
		return err
	}
	cr.l.digest = &digest

	cr.l.size = cr.count.n
	cr.l.consumed = true
	return nil
}

// countWriter counts bytes written to it.
type countWriter struct{ n int64 }

func (c *countWriter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

// multiCloser is a Closer that collects multiple Closers and Closes them in order.
type multiCloser []io.Closer

var _ io.Closer = (multiCloser)(nil)

func newMultiCloser(c ...io.Closer) multiCloser { return multiCloser(c) }

func (m multiCloser) Close() error {
	for _, c := range m {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}
