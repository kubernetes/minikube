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
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

type image struct {
	opener        Opener
	td            *tarDescriptor
	config        []byte
	imgDescriptor *singleImageTarDescriptor

	tag *name.Tag
}

type uncompressedImage struct {
	*image
}

type compressedImage struct {
	*image
	manifestLock sync.Mutex // Protects manifest
	manifest     *v1.Manifest
}

var _ partial.UncompressedImageCore = (*uncompressedImage)(nil)
var _ partial.CompressedImageCore = (*compressedImage)(nil)

// Opener is a thunk for opening a tar file.
type Opener func() (io.ReadCloser, error)

func pathOpener(path string) Opener {
	return func() (io.ReadCloser, error) {
		return os.Open(path)
	}
}

// ImageFromPath returns a v1.Image from a tarball located on path.
func ImageFromPath(path string, tag *name.Tag) (v1.Image, error) {
	return Image(pathOpener(path), tag)
}

// Image exposes an image from the tarball at the provided path.
func Image(opener Opener, tag *name.Tag) (v1.Image, error) {
	img := &image{
		opener: opener,
		tag:    tag,
	}
	if err := img.loadTarDescriptorAndConfig(); err != nil {
		return nil, err
	}

	// Peek at the first layer and see if it's compressed.
	compressed, err := img.areLayersCompressed()
	if err != nil {
		return nil, err
	}
	if compressed {
		c := compressedImage{
			image: img,
		}
		return partial.CompressedToImage(&c)
	}

	uc := uncompressedImage{
		image: img,
	}
	return partial.UncompressedToImage(&uc)
}

func (i *image) MediaType() (types.MediaType, error) {
	return types.DockerManifestSchema2, nil
}

// singleImageTarDescriptor is the struct used to represent a single image inside a `docker save` tarball.
type singleImageTarDescriptor struct {
	Config   string
	RepoTags []string
	Layers   []string
}

// tarDescriptor is the struct used inside the `manifest.json` file of a `docker save` tarball.
type tarDescriptor []singleImageTarDescriptor

func (td tarDescriptor) findSpecifiedImageDescriptor(tag *name.Tag) (*singleImageTarDescriptor, error) {
	if tag == nil {
		if len(td) != 1 {
			return nil, errors.New("tarball must contain only a single image to be used with tarball.Image")
		}
		return &(td)[0], nil
	}
	for _, img := range td {
		for _, tagStr := range img.RepoTags {
			repoTag, err := name.NewTag(tagStr, name.WeakValidation)
			if err != nil {
				return nil, err
			}

			// Compare the resolved names, since there are several ways to specify the same tag.
			if repoTag.Name() == tag.Name() {
				return &img, nil
			}
		}
	}
	return nil, fmt.Errorf("tag %s not found in tarball", tag)
}

func (i *image) areLayersCompressed() (bool, error) {
	if len(i.imgDescriptor.Layers) == 0 {
		return false, errors.New("0 layers found in image")
	}
	layer := i.imgDescriptor.Layers[0]
	blob, err := extractFileFromTar(i.opener, layer)
	if err != nil {
		return false, err
	}
	defer blob.Close()
	return v1util.IsGzipped(blob)
}

func (i *image) loadTarDescriptorAndConfig() error {
	td, err := extractFileFromTar(i.opener, "manifest.json")
	if err != nil {
		return err
	}
	defer td.Close()

	if err := json.NewDecoder(td).Decode(&i.td); err != nil {
		return err
	}

	i.imgDescriptor, err = i.td.findSpecifiedImageDescriptor(i.tag)
	if err != nil {
		return err
	}

	cfg, err := extractFileFromTar(i.opener, i.imgDescriptor.Config)
	if err != nil {
		return err
	}
	defer cfg.Close()

	i.config, err = ioutil.ReadAll(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (i *image) RawConfigFile() ([]byte, error) {
	return i.config, nil
}

// tarFile represents a single file inside a tar. Closing it closes the tar itself.
type tarFile struct {
	io.Reader
	io.Closer
}

func extractFileFromTar(opener Opener, filePath string) (io.ReadCloser, error) {
	f, err := opener()
	if err != nil {
		return nil, err
	}
	tf := tar.NewReader(f)
	for {
		hdr, err := tf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == filePath {
			return tarFile{
				Reader: tf,
				Closer: f,
			}, nil
		}
	}
	return nil, fmt.Errorf("file %s not found in tar", filePath)
}

// uncompressedLayerFromTarball implements partial.UncompressedLayer
type uncompressedLayerFromTarball struct {
	diffID   v1.Hash
	opener   Opener
	filePath string
}

// DiffID implements partial.UncompressedLayer
func (ulft *uncompressedLayerFromTarball) DiffID() (v1.Hash, error) {
	return ulft.diffID, nil
}

// Uncompressed implements partial.UncompressedLayer
func (ulft *uncompressedLayerFromTarball) Uncompressed() (io.ReadCloser, error) {
	return extractFileFromTar(ulft.opener, ulft.filePath)
}

func (i *uncompressedImage) LayerByDiffID(h v1.Hash) (partial.UncompressedLayer, error) {
	cfg, err := partial.ConfigFile(i)
	if err != nil {
		return nil, err
	}
	for idx, diffID := range cfg.RootFS.DiffIDs {
		if diffID == h {
			return &uncompressedLayerFromTarball{
				diffID:   diffID,
				opener:   i.opener,
				filePath: i.imgDescriptor.Layers[idx],
			}, nil
		}
	}
	return nil, fmt.Errorf("diff id %q not found", h)
}

func (c *compressedImage) Manifest() (*v1.Manifest, error) {
	c.manifestLock.Lock()
	defer c.manifestLock.Unlock()
	if c.manifest != nil {
		return c.manifest, nil
	}

	b, err := c.RawConfigFile()
	if err != nil {
		return nil, err
	}

	cfgHash, cfgSize, err := v1.SHA256(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	c.manifest = &v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.DockerManifestSchema2,
		Config: v1.Descriptor{
			MediaType: types.DockerConfigJSON,
			Size:      cfgSize,
			Digest:    cfgHash,
		},
	}

	for _, p := range c.imgDescriptor.Layers {
		l, err := extractFileFromTar(c.opener, p)
		if err != nil {
			return nil, err
		}
		defer l.Close()
		sha, size, err := v1.SHA256(l)
		if err != nil {
			return nil, err
		}
		c.manifest.Layers = append(c.manifest.Layers, v1.Descriptor{
			MediaType: types.DockerLayer,
			Size:      size,
			Digest:    sha,
		})
	}
	return c.manifest, nil
}

func (c *compressedImage) RawManifest() ([]byte, error) {
	return partial.RawManifest(c)
}

// compressedLayerFromTarball implements partial.CompressedLayer
type compressedLayerFromTarball struct {
	digest   v1.Hash
	opener   Opener
	filePath string
}

// Digest implements partial.CompressedLayer
func (clft *compressedLayerFromTarball) Digest() (v1.Hash, error) {
	return clft.digest, nil
}

// Compressed implements partial.CompressedLayer
func (clft *compressedLayerFromTarball) Compressed() (io.ReadCloser, error) {
	return extractFileFromTar(clft.opener, clft.filePath)
}

// Size implements partial.CompressedLayer
func (clft *compressedLayerFromTarball) Size() (int64, error) {
	r, err := clft.Compressed()
	if err != nil {
		return -1, err
	}
	defer r.Close()
	_, i, err := v1.SHA256(r)
	return i, err
}

func (c *compressedImage) LayerByDigest(h v1.Hash) (partial.CompressedLayer, error) {
	m, err := c.Manifest()
	if err != nil {
		return nil, err
	}
	for i, l := range m.Layers {
		if l.Digest == h {
			fp := c.imgDescriptor.Layers[i]
			return &compressedLayerFromTarball{
				digest:   h,
				opener:   c.opener,
				filePath: fp,
			}, nil
		}
	}
	return nil, fmt.Errorf("blob %v not found", h)
}
