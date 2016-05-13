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

// Package docker2aci implements a simple library for converting docker images to
// App Container Images (ACIs).
package docker2aci

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/appc/docker2aci/lib/backend/file"
	"github.com/appc/docker2aci/lib/backend/repository"
	"github.com/appc/docker2aci/lib/common"
	"github.com/appc/docker2aci/lib/types"
	"github.com/appc/docker2aci/lib/util"
	"github.com/appc/docker2aci/tarball"
	"github.com/appc/spec/pkg/acirenderer"
	"github.com/appc/spec/schema"
	appctypes "github.com/appc/spec/schema/types"
)

type Docker2ACIBackend interface {
	GetImageInfo(dockerUrl string) ([]string, *types.ParsedDockerURL, error)
	BuildACI(layerNumber int, layerID string, dockerURL *types.ParsedDockerURL, outputDir string, tmpBaseDir string, curPWl []string, compression common.Compression) (string, *schema.ImageManifest, error)
}

// Convert generates ACI images from docker registry URLs.
// It takes as input a dockerURL of the form:
//
// 	{docker registry URL}/{image name}:{tag}
//
// It then gets all the layers of the requested image and converts each of
// them to ACI.
// If the squash flag is true, it squashes all the layers in one file and
// places this file in outputDir; if it is false, it places every layer in its
// own ACI in outputDir.
// It will use the temporary directory specified by tmpDir, or the default
// temporary directory if tmpDir is "".
// username and password can be passed if the image needs authentication.
// It returns the list of generated ACI paths.
func Convert(dockerURL string, squash bool, outputDir string, tmpDir string, compression common.Compression, username string, password string, insecure bool) ([]string, error) {
	repositoryBackend := repository.NewRepositoryBackend(username, password, insecure)
	return convertReal(repositoryBackend, dockerURL, squash, outputDir, tmpDir, compression)
}

// ConvertFile generates ACI images from a file generated with "docker save".
// If there are several images/tags in the file, a particular image can be
// chosen with the syntax:
//
//	{docker registry URL}/{image name}:{tag}
//
// It takes as input the docker-generated file
//
// If the squash flag is true, it squashes all the layers in one file and
// places this file in outputDir; if it is false, it places every layer in its
// own ACI in outputDir.
// It will use the temporary directory specified by tmpDir, or the default
// temporary directory if tmpDir is "".
// It returns the list of generated ACI paths.
func ConvertFile(dockerURL string, filePath string, squash bool, outputDir string, tmpDir string, compression common.Compression) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	fileBackend := file.NewFileBackend(f)
	return convertReal(fileBackend, dockerURL, squash, outputDir, tmpDir, compression)
}

// GetIndexName returns the docker index server from a docker URL.
func GetIndexName(dockerURL string) string {
	index, _ := common.SplitReposName(dockerURL)
	return index
}

// GetDockercfgAuth reads a ~/.dockercfg file and returns the username and password
// of the given docker index server.
func GetDockercfgAuth(indexServer string) (string, string, error) {
	return common.GetAuthInfo(indexServer)
}

func convertReal(backend Docker2ACIBackend, dockerURL string, squash bool, outputDir string, tmpDir string, compression common.Compression) ([]string, error) {
	util.Debug("Getting image info...")
	ancestry, parsedDockerURL, err := backend.GetImageInfo(dockerURL)
	if err != nil {
		return nil, err
	}

	layersOutputDir := outputDir
	if squash {
		layersOutputDir, err = ioutil.TempDir(tmpDir, "docker2aci-")
		if err != nil {
			return nil, fmt.Errorf("error creating dir: %v", err)
		}
		defer os.RemoveAll(layersOutputDir)
	}

	conversionStore := NewConversionStore()

	var images acirenderer.Images
	var aciLayerPaths []string
	var curPwl []string
	for i := len(ancestry) - 1; i >= 0; i-- {
		layerID := ancestry[i]

		// only compress individual layers if we're not squashing
		layerCompression := compression
		if squash {
			layerCompression = common.NoCompression
		}

		aciPath, manifest, err := backend.BuildACI(i, layerID, parsedDockerURL, layersOutputDir, tmpDir, curPwl, layerCompression)
		if err != nil {
			return nil, fmt.Errorf("error building layer: %v", err)
		}

		key, err := conversionStore.WriteACI(aciPath)
		if err != nil {
			return nil, fmt.Errorf("error inserting in the conversion store: %v", err)
		}

		images = append(images, acirenderer.Image{Im: manifest, Key: key, Level: uint16(i)})
		aciLayerPaths = append(aciLayerPaths, aciPath)
		curPwl = manifest.PathWhitelist
	}

	// acirenderer expects images in order from upper to base layer
	images = util.ReverseImages(images)
	if squash {
		squashedImagePath, err := SquashLayers(images, conversionStore, *parsedDockerURL, outputDir, compression)
		if err != nil {
			return nil, fmt.Errorf("error squashing image: %v", err)
		}
		aciLayerPaths = []string{squashedImagePath}
	}

	return aciLayerPaths, nil
}

// SquashLayers receives a list of ACI layer file names ordered from base image
// to application image and squashes them into one ACI
func SquashLayers(images []acirenderer.Image, aciRegistry acirenderer.ACIRegistry, parsedDockerURL types.ParsedDockerURL, outputDir string, compression common.Compression) (path string, err error) {
	util.Debug("Squashing layers...")
	util.Debug("Rendering ACI...")
	renderedACI, err := acirenderer.GetRenderedACIFromList(images, aciRegistry)
	if err != nil {
		return "", fmt.Errorf("error rendering squashed image: %v", err)
	}
	manifests, err := getManifests(renderedACI, aciRegistry)
	if err != nil {
		return "", fmt.Errorf("error getting manifests: %v", err)
	}

	squashedFilename := getSquashedFilename(parsedDockerURL)
	squashedImagePath := filepath.Join(outputDir, squashedFilename)

	squashedTempFile, err := ioutil.TempFile(outputDir, "docker2aci-squashedFile-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err == nil {
			err = squashedTempFile.Close()
		} else {
			// remove temp file on error
			// we ignore its error to not mask the real error
			os.Remove(squashedTempFile.Name())
		}
	}()

	util.Debug("Writing squashed ACI...")
	if err := writeSquashedImage(squashedTempFile, renderedACI, aciRegistry, manifests, compression); err != nil {
		return "", fmt.Errorf("error writing squashed image: %v", err)
	}

	util.Debug("Validating squashed ACI...")
	if err := common.ValidateACI(squashedTempFile.Name()); err != nil {
		return "", fmt.Errorf("error validating image: %v", err)
	}

	if err := os.Rename(squashedTempFile.Name(), squashedImagePath); err != nil {
		return "", err
	}

	util.Debug("ACI squashed!")
	return squashedImagePath, nil
}

func getSquashedFilename(parsedDockerURL types.ParsedDockerURL) string {
	squashedFilename := strings.Replace(parsedDockerURL.ImageName, "/", "-", -1)
	if parsedDockerURL.Tag != "" {
		squashedFilename += "-" + parsedDockerURL.Tag
	}
	squashedFilename += ".aci"

	return squashedFilename
}

func getManifests(renderedACI acirenderer.RenderedACI, aciRegistry acirenderer.ACIRegistry) ([]schema.ImageManifest, error) {
	var manifests []schema.ImageManifest

	for _, aci := range renderedACI {
		im, err := aciRegistry.GetImageManifest(aci.Key)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, *im)
	}

	return manifests, nil
}

func writeSquashedImage(outputFile *os.File, renderedACI acirenderer.RenderedACI, aciProvider acirenderer.ACIProvider, manifests []schema.ImageManifest, compression common.Compression) error {
	var tarWriterTarget io.WriteCloser = outputFile

	switch compression {
	case common.NoCompression:
	case common.GzipCompression:
		tarWriterTarget = gzip.NewWriter(outputFile)
		defer tarWriterTarget.Close()
	default:
		return fmt.Errorf("unexpected compression enum value: %d", compression)
	}

	outputWriter := tar.NewWriter(tarWriterTarget)
	defer outputWriter.Close()

	finalManifest := mergeManifests(manifests)

	if err := common.WriteManifest(outputWriter, finalManifest); err != nil {
		return err
	}

	if err := common.WriteRootfsDir(outputWriter); err != nil {
		return err
	}

	type hardLinkEntry struct {
		firstLinkCleanName string
		firstLinkHeader    tar.Header
		keepOriginal       bool
		walked             bool
	}
	// map aciFileKey -> cleanTarget -> hardLinkEntry
	hardLinks := make(map[string]map[string]hardLinkEntry)

	// first pass: read all the entries and build the hardLinks map in memory
	// but don't write on disk
	for _, aciFile := range renderedACI {
		rs, err := aciProvider.ReadStream(aciFile.Key)
		if err != nil {
			return err
		}
		defer rs.Close()

		hardLinks[aciFile.Key] = map[string]hardLinkEntry{}

		squashWalker := func(t *tarball.TarFile) error {
			cleanName := filepath.Clean(t.Name())
			// the rootfs and the squashed manifest are added separately
			if cleanName == "manifest" || cleanName == "rootfs" {
				return nil
			}
			_, keep := aciFile.FileMap[cleanName]
			if keep && t.Header.Typeflag == tar.TypeLink {
				cleanTarget := filepath.Clean(t.Linkname())
				if _, ok := hardLinks[aciFile.Key][cleanTarget]; !ok {
					_, keepOriginal := aciFile.FileMap[cleanTarget]
					hardLinks[aciFile.Key][cleanTarget] = hardLinkEntry{cleanName, *t.Header, keepOriginal, false}
				}
			}
			return nil
		}

		tr := tar.NewReader(rs)
		if err := tarball.Walk(*tr, squashWalker); err != nil {
			return err
		}
	}

	// second pass: write on disk
	for _, aciFile := range renderedACI {
		rs, err := aciProvider.ReadStream(aciFile.Key)
		if err != nil {
			return err
		}
		defer rs.Close()

		squashWalker := func(t *tarball.TarFile) error {
			cleanName := filepath.Clean(t.Name())
			// the rootfs and the squashed manifest are added separately
			if cleanName == "manifest" || cleanName == "rootfs" {
				return nil
			}
			_, keep := aciFile.FileMap[cleanName]

			if link, ok := hardLinks[aciFile.Key][cleanName]; ok {
				if keep != link.keepOriginal {
					return fmt.Errorf("logic error: should we keep file %q?", cleanName)
				}
				if keep {
					if err := outputWriter.WriteHeader(t.Header); err != nil {
						return fmt.Errorf("error writing header: %v", err)
					}
					if _, err := io.Copy(outputWriter, t.TarStream); err != nil {
						return fmt.Errorf("error copying file into the tar out: %v", err)
					}
				} else {
					// The current file does not remain but there is a hard link pointing to
					// it. Write the current file but with the filename of the first hard link
					// pointing to it. That first hard link will not be written later, see
					// variable "alreadyWritten".
					link.firstLinkHeader.Size = t.Header.Size
					link.firstLinkHeader.Typeflag = t.Header.Typeflag
					link.firstLinkHeader.Linkname = ""

					if err := outputWriter.WriteHeader(&link.firstLinkHeader); err != nil {
						return fmt.Errorf("error writing header: %v", err)
					}
					if _, err := io.Copy(outputWriter, t.TarStream); err != nil {
						return fmt.Errorf("error copying file into the tar out: %v", err)
					}
				}
			} else if keep {
				alreadyWritten := false
				if t.Header.Typeflag == tar.TypeLink {
					cleanTarget := filepath.Clean(t.Linkname())
					if link, ok := hardLinks[aciFile.Key][cleanTarget]; ok {
						if !link.keepOriginal {
							if link.walked {
								t.Header.Linkname = link.firstLinkCleanName
							} else {
								alreadyWritten = true
							}
						}
						link.walked = true
						hardLinks[aciFile.Key][cleanTarget] = link
					}
				}

				if !alreadyWritten {
					if err := outputWriter.WriteHeader(t.Header); err != nil {
						return fmt.Errorf("error writing header: %v", err)
					}
					if _, err := io.Copy(outputWriter, t.TarStream); err != nil {
						return fmt.Errorf("error copying file into the tar out: %v", err)
					}
				}
			}
			return nil
		}

		tr := tar.NewReader(rs)
		if err := tarball.Walk(*tr, squashWalker); err != nil {
			return err
		}
	}

	return nil
}

func mergeManifests(manifests []schema.ImageManifest) schema.ImageManifest {
	// FIXME(iaguis) we take app layer's manifest as the final manifest for now
	manifest := manifests[0]

	manifest.Dependencies = nil

	layerIndex := -1
	for i, l := range manifest.Labels {
		if l.Name.String() == "layer" {
			layerIndex = i
		}
	}

	if layerIndex != -1 {
		manifest.Labels = append(manifest.Labels[:layerIndex], manifest.Labels[layerIndex+1:]...)
	}

	nameWithoutLayerID := appctypes.MustACIdentifier(stripLayerID(manifest.Name.String()))

	manifest.Name = *nameWithoutLayerID

	// once the image is squashed, we don't need a pathWhitelist
	manifest.PathWhitelist = nil

	return manifest
}

// striplayerID strips the layer ID from an app name:
//
// myregistry.com/organization/app-name-85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281
// myregistry.com/organization/app-name
func stripLayerID(layerName string) string {
	n := strings.LastIndex(layerName, "-")
	return layerName[:n]
}
