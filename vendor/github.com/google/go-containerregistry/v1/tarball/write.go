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
	"io"
	"os"

	"github.com/google/go-containerregistry/name"
	"github.com/google/go-containerregistry/v1"
)

// WriteOptions are used to expose optional information to guide or
// control the image write.
type WriteOptions struct {
	// TODO(mattmoor): Whether to store things compressed?
}

// Write saves the image as the given tag in a tarball at the given path.
func Write(p string, tag name.Tag, img v1.Image, wo *WriteOptions) error {
	// Write in the compressed format.
	// This is a tarball, on-disk, with:
	// One manifest.json file at the top level containing information about several images.
	// One file for each layer, named after the layer's SHA.
	// One file for the config blob, named after its SHA.

	w, err := os.OpenFile(p, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer w.Close()

	tf := tar.NewWriter(w)
	defer tf.Close()

	// Write the config.
	cfgName, err := img.ConfigName()
	if err != nil {
		return err
	}
	cfg, err := img.ConfigFile()
	if err != nil {
		return err
	}
	cfgBlob, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := writeFile(tf, cfgName.String(), bytes.NewReader(cfgBlob), int64(len(cfgBlob))); err != nil {
		return err
	}

	// Write the layers.
	layers, err := img.Layers()
	if err != nil {
		return err
	}
	layerPaths := []string{}
	for _, l := range layers {
		d, err := l.Digest()
		if err != nil {
			return err
		}
		layerPaths = append(layerPaths, d.String())
		r, err := l.Compressed()
		if err != nil {
			return err
		}
		blobSize, err := l.Size()
		if err != nil {
			return err
		}

		if err := writeFile(tf, d.String(), r, blobSize); err != nil {
			return err
		}
	}

	// Generate the tar descriptor and write it.
	td := tarDescriptor{
		singleImageTarDescriptor{
			Config:   cfgName.String(),
			RepoTags: []string{tag.String()},
			Layers:   layerPaths,
		},
	}
	tdBytes, err := json.Marshal(td)
	if err != nil {
		return err
	}
	return writeFile(tf, "manifest.json", bytes.NewReader(tdBytes), int64(len(tdBytes)))
}

func writeFile(tf *tar.Writer, path string, r io.Reader, size int64) error {
	hdr := &tar.Header{
		Mode:     0644,
		Typeflag: tar.TypeReg,
		Size:     size,
		Name:     path,
	}
	if err := tf.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := io.Copy(tf, r)
	return err
}
