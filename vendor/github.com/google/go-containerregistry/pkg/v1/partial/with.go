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

package partial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

// WithRawConfigFile defines the subset of v1.Image used by these helper methods
type WithRawConfigFile interface {
	// RawConfigFile returns the serialized bytes of this image's config file.
	RawConfigFile() ([]byte, error)
}

// ConfigFile is a helper for implementing v1.Image
func ConfigFile(i WithRawConfigFile) (*v1.ConfigFile, error) {
	b, err := i.RawConfigFile()
	if err != nil {
		return nil, err
	}
	return v1.ParseConfigFile(bytes.NewReader(b))
}

// ConfigName is a helper for implementing v1.Image
func ConfigName(i WithRawConfigFile) (v1.Hash, error) {
	b, err := i.RawConfigFile()
	if err != nil {
		return v1.Hash{}, err
	}
	h, _, err := v1.SHA256(bytes.NewReader(b))
	return h, err
}

type configLayer struct {
	hash    v1.Hash
	content []byte
}

// Digest implements v1.Layer
func (cl *configLayer) Digest() (v1.Hash, error) {
	return cl.hash, nil
}

// DiffID implements v1.Layer
func (cl *configLayer) DiffID() (v1.Hash, error) {
	return cl.hash, nil
}

// Uncompressed implements v1.Layer
func (cl *configLayer) Uncompressed() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer(cl.content)), nil
}

// Compressed implements v1.Layer
func (cl *configLayer) Compressed() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer(cl.content)), nil
}

// Size implements v1.Layer
func (cl *configLayer) Size() (int64, error) {
	return int64(len(cl.content)), nil
}

var _ v1.Layer = (*configLayer)(nil)

// ConfigLayer implements v1.Layer from the raw config bytes.
// This is so that clients (e.g. remote) can access the config as a blob.
func ConfigLayer(i WithRawConfigFile) (v1.Layer, error) {
	h, err := ConfigName(i)
	if err != nil {
		return nil, err
	}
	rcfg, err := i.RawConfigFile()
	if err != nil {
		return nil, err
	}
	return &configLayer{
		hash:    h,
		content: rcfg,
	}, nil
}

// WithConfigFile defines the subset of v1.Image used by these helper methods
type WithConfigFile interface {
	// ConfigFile returns this image's config file.
	ConfigFile() (*v1.ConfigFile, error)
}

// DiffIDs is a helper for implementing v1.Image
func DiffIDs(i WithConfigFile) ([]v1.Hash, error) {
	cfg, err := i.ConfigFile()
	if err != nil {
		return nil, err
	}
	return cfg.RootFS.DiffIDs, nil
}

// RawConfigFile is a helper for implementing v1.Image
func RawConfigFile(i WithConfigFile) ([]byte, error) {
	cfg, err := i.ConfigFile()
	if err != nil {
		return nil, err
	}
	return json.Marshal(cfg)
}

// WithUncompressedLayer defines the subset of v1.Image used by these helper methods
type WithUncompressedLayer interface {
	// UncompressedLayer is like UncompressedBlob, but takes the "diff id".
	UncompressedLayer(v1.Hash) (io.ReadCloser, error)
}

// Layer is the same as Blob, but takes the "diff id".
func Layer(wul WithUncompressedLayer, h v1.Hash) (io.ReadCloser, error) {
	rc, err := wul.UncompressedLayer(h)
	if err != nil {
		return nil, err
	}
	return v1util.GzipReadCloser(rc)
}

// WithRawManifest defines the subset of v1.Image used by these helper methods
type WithRawManifest interface {
	// RawManifest returns the serialized bytes of this image's config file.
	RawManifest() ([]byte, error)
}

// Digest is a helper for implementing v1.Image
func Digest(i WithRawManifest) (v1.Hash, error) {
	mb, err := i.RawManifest()
	if err != nil {
		return v1.Hash{}, err
	}
	digest, _, err := v1.SHA256(bytes.NewReader(mb))
	return digest, err
}

// Manifest is a helper for implementing v1.Image
func Manifest(i WithRawManifest) (*v1.Manifest, error) {
	b, err := i.RawManifest()
	if err != nil {
		return nil, err
	}
	return v1.ParseManifest(bytes.NewReader(b))
}

// WithManifest defines the subset of v1.Image used by these helper methods
type WithManifest interface {
	// Manifest returns this image's Manifest object.
	Manifest() (*v1.Manifest, error)
}

// RawManifest is a helper for implementing v1.Image
func RawManifest(i WithManifest) ([]byte, error) {
	m, err := i.Manifest()
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// FSLayers is a helper for implementing v1.Image
func FSLayers(i WithManifest) ([]v1.Hash, error) {
	m, err := i.Manifest()
	if err != nil {
		return nil, err
	}
	fsl := make([]v1.Hash, len(m.Layers))
	for i, l := range m.Layers {
		fsl[i] = l.Digest
	}
	return fsl, nil
}

// BlobSize is a helper for implementing v1.Image
func BlobSize(i WithManifest, h v1.Hash) (int64, error) {
	m, err := i.Manifest()
	if err != nil {
		return -1, err
	}
	for _, l := range m.Layers {
		if l.Digest == h {
			return l.Size, nil
		}
	}
	return -1, fmt.Errorf("blob %v not found", h)
}

// WithManifestAndConfigFile defines the subset of v1.Image used by these helper methods
type WithManifestAndConfigFile interface {
	WithConfigFile

	// Manifest returns this image's Manifest object.
	Manifest() (*v1.Manifest, error)
}

// BlobToDiffID is a helper for mapping between compressed
// and uncompressed blob hashes.
func BlobToDiffID(i WithManifestAndConfigFile, h v1.Hash) (v1.Hash, error) {
	blobs, err := FSLayers(i)
	if err != nil {
		return v1.Hash{}, err
	}
	diffIDs, err := DiffIDs(i)
	if err != nil {
		return v1.Hash{}, err
	}
	if len(blobs) != len(diffIDs) {
		return v1.Hash{}, fmt.Errorf("mismatched fs layers (%d) and diff ids (%d)", len(blobs), len(diffIDs))
	}
	for i, blob := range blobs {
		if blob == h {
			return diffIDs[i], nil
		}
	}
	return v1.Hash{}, fmt.Errorf("unknown blob %v", h)
}

// DiffIDToBlob is a helper for mapping between uncompressed
// and compressed blob hashes.
func DiffIDToBlob(wm WithManifestAndConfigFile, h v1.Hash) (v1.Hash, error) {
	blobs, err := FSLayers(wm)
	if err != nil {
		return v1.Hash{}, err
	}
	diffIDs, err := DiffIDs(wm)
	if err != nil {
		return v1.Hash{}, err
	}
	if len(blobs) != len(diffIDs) {
		return v1.Hash{}, fmt.Errorf("mismatched fs layers (%d) and diff ids (%d)", len(blobs), len(diffIDs))
	}
	for i, diffID := range diffIDs {
		if diffID == h {
			return blobs[i], nil
		}
	}
	return v1.Hash{}, fmt.Errorf("unknown diffID %v", h)

}

// WithBlob defines the subset of v1.Image used by these helper methods
type WithBlob interface {
	// Blob returns a ReadCloser for streaming the blob's content.
	Blob(v1.Hash) (io.ReadCloser, error)
}

// UncompressedBlob returns a ReadCloser for streaming the blob's content uncompressed.
func UncompressedBlob(b WithBlob, h v1.Hash) (io.ReadCloser, error) {
	rc, err := b.Blob(h)
	if err != nil {
		return nil, err
	}
	return v1util.GunzipReadCloser(rc)
}

// WithDiffID defines the subset of v1.Layer for exposing the DiffID method.
type WithDiffID interface {
	DiffID() (v1.Hash, error)
}
