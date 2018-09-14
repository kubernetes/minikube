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
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

// remoteImage accesses an image from a remote registry
type remoteImage struct {
	ref          name.Reference
	client       *http.Client
	manifestLock sync.Mutex // Protects manifest
	manifest     []byte
	configLock   sync.Mutex // Protects config
	config       []byte
}

type ImageOption func(*imageOpener) error

var _ partial.CompressedImageCore = (*remoteImage)(nil)

type imageOpener struct {
	auth      authn.Authenticator
	transport http.RoundTripper
	ref       name.Reference
	client    *http.Client
}

func (i *imageOpener) Open() (v1.Image, error) {
	tr, err := transport.New(i.ref.Context().Registry, i.auth, i.transport, []string{i.ref.Scope(transport.PullScope)})
	if err != nil {
		return nil, err
	}
	ri := &remoteImage{
		ref:    i.ref,
		client: &http.Client{Transport: tr},
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
	}

	for _, option := range options {
		if err := option(img); err != nil {
			return nil, err
		}
	}
	return img.Open()
}

func (r *remoteImage) url(resource, identifier string) url.URL {
	return url.URL{
		Scheme: r.ref.Context().Registry.Scheme(),
		Host:   r.ref.Context().RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s/%s/%s", r.ref.Context().RepositoryStr(), resource, identifier),
	}
}

func (r *remoteImage) MediaType() (types.MediaType, error) {
	// TODO(jonjohnsonjr): Determine this based on response.
	return types.DockerManifestSchema2, nil
}

// TODO(jonjohnsonjr): Handle manifest lists.
func (r *remoteImage) RawManifest() ([]byte, error) {
	r.manifestLock.Lock()
	defer r.manifestLock.Unlock()
	if r.manifest != nil {
		return r.manifest, nil
	}

	u := r.url("manifests", r.ref.Identifier())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	// TODO(jonjohnsonjr): Accept OCI manifest, manifest list, and image index.
	req.Header.Set("Accept", string(types.DockerManifestSchema2))
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	digest, _, err := v1.SHA256(bytes.NewReader(manifest))
	if err != nil {
		return nil, err
	}

	// Validate the digest matches what we asked for, if pulling by digest.
	if dgst, ok := r.ref.(name.Digest); ok {
		if digest.String() != dgst.DigestStr() {
			return nil, fmt.Errorf("manifest digest: %q does not match requested digest: %q for %q", digest, dgst.DigestStr(), r.ref)
		}
	} else if checksum := resp.Header.Get("Docker-Content-Digest"); checksum != "" && checksum != digest.String() {
		err := fmt.Errorf("manifest digest: %q does not match Docker-Content-Digest: %q for %q", digest, checksum, r.ref)
		if r.ref.Context().RegistryStr() == name.DefaultRegistry {
			// TODO(docker/distribution#2395): Remove this check.
		} else {
			// When pulling by tag, we can only validate that the digest matches what the registry told us it should be.
			return nil, err
		}
	}

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
	resp, err := rl.ri.client.Get(u.String())
	if err != nil {
		return nil, err
	}

	if err := CheckError(resp, http.StatusOK); err != nil {
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
