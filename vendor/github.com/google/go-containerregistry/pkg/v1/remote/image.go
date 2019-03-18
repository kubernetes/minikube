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

package remote

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

var defaultPlatform = v1.Platform{
	Architecture: "amd64",
	OS:           "linux",
}

// remoteImage accesses an image from a remote registry
type remoteImage struct {
	fetcher
	manifestLock sync.Mutex // Protects manifest
	manifest     []byte
	configLock   sync.Mutex // Protects config
	config       []byte
	mediaType    types.MediaType
	platform     v1.Platform
}

// ImageOption is a functional option for Image.
type ImageOption func(*imageOpener) error

var _ partial.CompressedImageCore = (*remoteImage)(nil)

type imageOpener struct {
	auth      authn.Authenticator
	transport http.RoundTripper
	ref       name.Reference
	client    *http.Client
	platform  v1.Platform
}

func (i *imageOpener) Open() (v1.Image, error) {
	tr, err := transport.New(i.ref.Context().Registry, i.auth, i.transport, []string{i.ref.Scope(transport.PullScope)})
	if err != nil {
		return nil, err
	}
	ri := &remoteImage{
		fetcher: fetcher{
			Ref:    i.ref,
			Client: &http.Client{Transport: tr},
		},
		platform: i.platform,
	}
	imgCore, err := partial.CompressedToImage(ri)
	if err != nil {
		return imgCore, err
	}
	// Wrap the v1.Layers returned by this v1.Image in a hint for downstream
	// remote.Write calls to facilitate cross-repo "mounting".
	return &mountableImage{
		Image:     imgCore,
		Reference: i.ref,
	}, nil
}

// Image provides access to a remote image reference, applying functional options
// to the underlying imageOpener before resolving the reference into a v1.Image.
func Image(ref name.Reference, options ...ImageOption) (v1.Image, error) {
	img := &imageOpener{
		auth:      authn.Anonymous,
		transport: http.DefaultTransport,
		ref:       ref,
		platform:  defaultPlatform,
	}

	for _, option := range options {
		if err := option(img); err != nil {
			return nil, err
		}
	}
	return img.Open()
}

// fetcher implements methods for reading from a remote image.
type fetcher struct {
	Ref    name.Reference
	Client *http.Client
}

// url returns a url.Url for the specified path in the context of this remote image reference.
func (f *fetcher) url(resource, identifier string) url.URL {
	return url.URL{
		Scheme: f.Ref.Context().Registry.Scheme(),
		Host:   f.Ref.Context().RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s/%s/%s", f.Ref.Context().RepositoryStr(), resource, identifier),
	}
}

func (f *fetcher) fetchManifest(acceptable []types.MediaType) ([]byte, *v1.Descriptor, error) {
	u := f.url("manifests", f.Ref.Identifier())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	accept := []string{}
	for _, mt := range acceptable {
		accept = append(accept, string(mt))
	}
	req.Header.Set("Accept", strings.Join(accept, ","))

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		return nil, nil, err
	}

	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	digest, size, err := v1.SHA256(bytes.NewReader(manifest))
	if err != nil {
		return nil, nil, err
	}

	// Validate the digest matches what we asked for, if pulling by digest.
	if dgst, ok := f.Ref.(name.Digest); ok {
		if digest.String() != dgst.DigestStr() {
			return nil, nil, fmt.Errorf("manifest digest: %q does not match requested digest: %q for %q", digest, dgst.DigestStr(), f.Ref)
		}
	} else {
		// Do nothing for tags; I give up.
		//
		// We'd like to validate that the "Docker-Content-Digest" header matches what is returned by the registry,
		// but so many registries implement this incorrectly that it's not worth checking.
		//
		// For reference:
		// https://github.com/docker/distribution/issues/2395
		// https://github.com/GoogleContainerTools/kaniko/issues/298
	}

	// Return all this info since we have to calculate it anyway.
	desc := v1.Descriptor{
		Digest:    digest,
		Size:      size,
		MediaType: types.MediaType(resp.Header.Get("Content-Type")),
	}

	return manifest, &desc, nil
}

func (r *remoteImage) MediaType() (types.MediaType, error) {
	if string(r.mediaType) != "" {
		return r.mediaType, nil
	}
	return types.DockerManifestSchema2, nil
}

// TODO(jonjohnsonjr): Handle manifest lists.
func (r *remoteImage) RawManifest() ([]byte, error) {
	r.manifestLock.Lock()
	defer r.manifestLock.Unlock()
	if r.manifest != nil {
		return r.manifest, nil
	}

	acceptable := []types.MediaType{
		types.DockerManifestSchema2,
		types.OCIManifestSchema1,
		// We'll resolve these to an image based on the platform.
		types.DockerManifestList,
		types.OCIImageIndex,
	}
	manifest, desc, err := r.fetchManifest(acceptable)
	if err != nil {
		return nil, err
	}

	// We want an image but the registry has an index, resolve it to an image.
	for desc.MediaType == types.DockerManifestList || desc.MediaType == types.OCIImageIndex {
		manifest, desc, err = r.matchImage(manifest)
		if err != nil {
			return nil, err
		}
	}

	r.mediaType = desc.MediaType
	r.manifest = manifest
	return r.manifest, nil
}

func (r *remoteImage) RawConfigFile() ([]byte, error) {
	r.configLock.Lock()
	defer r.configLock.Unlock()
	if r.config != nil {
		return r.config, nil
	}

	m, err := partial.Manifest(r)
	if err != nil {
		return nil, err
	}

	cl, err := r.LayerByDigest(m.Config.Digest)
	if err != nil {
		return nil, err
	}
	body, err := cl.Compressed()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	r.config, err = ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return r.config, nil
}

// remoteLayer implements partial.CompressedLayer
type remoteLayer struct {
	ri     *remoteImage
	digest v1.Hash
}

// Digest implements partial.CompressedLayer
func (rl *remoteLayer) Digest() (v1.Hash, error) {
	return rl.digest, nil
}

// Compressed implements partial.CompressedLayer
func (rl *remoteLayer) Compressed() (io.ReadCloser, error) {
	u := rl.ri.url("blobs", rl.digest.String())
	resp, err := rl.ri.Client.Get(u.String())
	if err != nil {
		return nil, err
	}

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return v1util.VerifyReadCloser(resp.Body, rl.digest)
}

// Manifest implements partial.WithManifest so that we can use partial.BlobSize below.
func (rl *remoteLayer) Manifest() (*v1.Manifest, error) {
	return partial.Manifest(rl.ri)
}

// Size implements partial.CompressedLayer
func (rl *remoteLayer) Size() (int64, error) {
	// Look up the size of this digest in the manifest to avoid a request.
	return partial.BlobSize(rl, rl.digest)
}

// ConfigFile implements partial.WithManifestAndConfigFile so that we can use partial.BlobToDiffID below.
func (rl *remoteLayer) ConfigFile() (*v1.ConfigFile, error) {
	return partial.ConfigFile(rl.ri)
}

// DiffID implements partial.WithDiffID so that we don't recompute a DiffID that we already have
// available in our ConfigFile.
func (rl *remoteLayer) DiffID() (v1.Hash, error) {
	return partial.BlobToDiffID(rl, rl.digest)
}

// LayerByDigest implements partial.CompressedLayer
func (r *remoteImage) LayerByDigest(h v1.Hash) (partial.CompressedLayer, error) {
	return &remoteLayer{
		ri:     r,
		digest: h,
	}, nil
}

// This naively matches the first manifest with matching Architecture and OS.
//
// We should probably use this instead:
//	 github.com/containerd/containerd/platforms
//
// But first we'd need to migrate to:
//   github.com/opencontainers/image-spec/specs-go/v1
func (r *remoteImage) matchImage(rawIndex []byte) ([]byte, *v1.Descriptor, error) {
	index, err := v1.ParseIndexManifest(bytes.NewReader(rawIndex))
	if err != nil {
		return nil, nil, err
	}
	for _, childDesc := range index.Manifests {
		// If platform is missing from child descriptor, assume it's amd64/linux.
		p := defaultPlatform
		if childDesc.Platform != nil {
			p = *childDesc.Platform
		}
		if r.platform.Architecture == p.Architecture && r.platform.OS == p.OS {
			childRef, err := name.ParseReference(fmt.Sprintf("%s@%s", r.Ref.Context(), childDesc.Digest), name.StrictValidation)
			if err != nil {
				return nil, nil, err
			}
			r.fetcher = fetcher{
				Client: r.Client,
				Ref:    childRef,
			}
			return r.fetchManifest([]types.MediaType{childDesc.MediaType})
		}
	}
	return nil, nil, fmt.Errorf("no matching image for %s/%s, index: %s", r.platform.Architecture, r.platform.OS, string(rawIndex))
}
