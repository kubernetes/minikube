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
	"fmt"
	"io"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// WriteToFile writes in the compressed format to a tarball, on disk.
// This is just syntactic sugar wrapping tarball.Write with a new file.
func WriteToFile(p string, ref name.Reference, img v1.Image) error {
	w, err := os.Create(p)
	if err != nil {
		return err
	}
	defer w.Close()

	return Write(ref, img, w)
}

// MultiWriteToFile writes in the compressed format to a tarball, on disk.
// This is just syntactic sugar wrapping tarball.MultiWrite with a new file.
func MultiWriteToFile(p string, tagToImage map[name.Tag]v1.Image) error {
	var refToImage map[name.Reference]v1.Image = make(map[name.Reference]v1.Image, len(tagToImage))
	for i, d := range tagToImage {
		refToImage[i] = d
	}
	return MultiRefWriteToFile(p, refToImage)
}

// MultiRefWriteToFile writes in the compressed format to a tarball, on disk.
// This is just syntactic sugar wrapping tarball.MultiRefWrite with a new file.
func MultiRefWriteToFile(p string, refToImage map[name.Reference]v1.Image) error {
	w, err := os.Create(p)
	if err != nil {
		return err
	}
	defer w.Close()

	return MultiRefWrite(refToImage, w)
}

// Write is a wrapper to write a single image and tag to a tarball.
func Write(ref name.Reference, img v1.Image, w io.Writer) error {
	return MultiRefWrite(map[name.Reference]v1.Image{ref: img}, w)
}

// MultiWrite writes the contents of each image to the provided reader, in the compressed format.
// The contents are written in the following format:
// One manifest.json file at the top level containing information about several images.
// One file for each layer, named after the layer's SHA.
// One file for the config blob, named after its SHA.
func MultiWrite(tagToImage map[name.Tag]v1.Image, w io.Writer) error {
	var refToImage map[name.Reference]v1.Image = make(map[name.Reference]v1.Image, len(tagToImage))
	for i, d := range tagToImage {
		refToImage[i] = d
	}
	return MultiRefWrite(refToImage, w)
}

// MultiRefWrite writes the contents of each image to the provided reader, in the compressed format.
// The contents are written in the following format:
// One manifest.json file at the top level containing information about several images.
// One file for each layer, named after the layer's SHA.
// One file for the config blob, named after its SHA.
func MultiRefWrite(refToImage map[name.Reference]v1.Image, w io.Writer) error {
	tf := tar.NewWriter(w)
	defer tf.Close()

	imageToTags := dedupRefToImage(refToImage)
	var td tarDescriptor

	for img, tags := range imageToTags {
		// Write the config.
		cfgName, err := img.ConfigName()
		if err != nil {
			return err
		}
		cfgBlob, err := img.RawConfigFile()
		if err != nil {
			return err
		}
		if err := writeTarEntry(tf, cfgName.String(), bytes.NewReader(cfgBlob), int64(len(cfgBlob))); err != nil {
			return err
		}

		// Write the layers.
		layers, err := img.Layers()
		if err != nil {
			return err
		}
		layerFiles := make([]string, len(layers))
		for i, l := range layers {
			d, err := l.Digest()
			if err != nil {
				return err
			}

			// Munge the file name to appease ancient technology.
			//
			// tar assumes anything with a colon is a remote tape drive:
			// https://www.gnu.org/software/tar/manual/html_section/tar_45.html
			// Drop the algorithm prefix, e.g. "sha256:"
			hex := d.Hex

			// gunzip expects certain file extensions:
			// https://www.gnu.org/software/gzip/manual/html_node/Overview.html
			layerFiles[i] = fmt.Sprintf("%s.tar.gz", hex)

			r, err := l.Compressed()
			if err != nil {
				return err
			}
			blobSize, err := l.Size()
			if err != nil {
				return err
			}

			if err := writeTarEntry(tf, layerFiles[i], r, blobSize); err != nil {
				return err
			}
		}

		// Generate the tar descriptor and write it.
		sitd := singleImageTarDescriptor{
			Config:   cfgName.String(),
			RepoTags: tags,
			Layers:   layerFiles,
		}

		td = append(td, sitd)
	}

	tdBytes, err := json.Marshal(td)
	if err != nil {
		return err
	}
	return writeTarEntry(tf, "manifest.json", bytes.NewReader(tdBytes), int64(len(tdBytes)))
}

func dedupRefToImage(refToImage map[name.Reference]v1.Image) map[v1.Image][]string {
	imageToTags := make(map[v1.Image][]string)

	for ref, img := range refToImage {
		if tag, ok := ref.(name.Tag); ok {
			if tags, ok := imageToTags[img]; ok && tags != nil {
				imageToTags[img] = append(tags, tag.String())
			} else {
				imageToTags[img] = []string{tag.String()}
			}
		} else {
			if _, ok := imageToTags[img]; !ok {
				imageToTags[img] = nil
			}
		}
	}

	return imageToTags
}

// write a file to the provided writer with a corresponding tar header
func writeTarEntry(tf *tar.Writer, path string, r io.Reader, size int64) error {
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
