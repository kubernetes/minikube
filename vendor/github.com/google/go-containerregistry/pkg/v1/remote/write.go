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
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// WriteOptions are used to expose optional information to guide or
// control the image write.
type WriteOptions struct {
	// TODO(mattmoor): Expose "threads" to limit parallelism?
}

// Write pushes the provided img to the specified image reference.
func Write(ref name.Reference, img v1.Image, auth authn.Authenticator, t http.RoundTripper,
	wo WriteOptions) error {

	ls, err := img.Layers()
	if err != nil {
		return err
	}

	scopes := scopesForUploadingImage(ref, ls)
	tr, err := transport.New(ref.Context().Registry, auth, t, scopes)
	if err != nil {
		return err
	}
	w := writer{
		ref:     ref,
		client:  &http.Client{Transport: tr},
		img:     img,
		options: wo,
	}

	bs, err := img.BlobSet()
	if err != nil {
		return err
	}

	// Spin up go routines to publish each of the members of BlobSet(),
	// and use an error channel to collect their results.
	errCh := make(chan error)
	defer close(errCh)
	for h := range bs {
		go func(h v1.Hash) {
			errCh <- w.uploadOne(h)
		}(h)
	}

	// Now wait for all of the blob uploads to complete.
	var errors []error
	for _ = range bs {
		if err := <-errCh; err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		// Return the first error we encountered.
		return errors[0]
	}

	// With all of the constituent elements uploaded, upload the manifest
	// to commit the image.
	return w.commitImage()
}

// writer writes the elements of an image to a remote image reference.
type writer struct {
	ref     name.Reference
	client  *http.Client
	img     v1.Image
	options WriteOptions
}

// url returns a url.Url for the specified path in the context of this remote image reference.
func (w *writer) url(path string) url.URL {
	return url.URL{
		Scheme: w.ref.Context().Registry.Scheme(),
		Host:   w.ref.Context().RegistryStr(),
		Path:   path,
	}
}

// nextLocation extracts the fully-qualified URL to which we should send the next request in an upload sequence.
func (w *writer) nextLocation(resp *http.Response) (string, error) {
	loc := resp.Header.Get("Location")
	if len(loc) == 0 {
		return "", errors.New("missing Location header")
	}
	u, err := url.Parse(loc)
	if err != nil {
		return "", err
	}

	// If the location header returned is just a url path, then fully qualify it.
	// We cannot simply call w.url, since there might be an embedded query string.
	return resp.Request.URL.ResolveReference(u).String(), nil
}

// checkExisting checks if a blob exists already in the repository by making a
// HEAD request to the blob store API.  GCR performs an existence check on the
// initiation if "mount" is specified, even if no "from" sources are specified.
// However, this is not broadly applicable to all registries, e.g. ECR.
func (w *writer) checkExisting(h v1.Hash) (bool, error) {
	u := w.url(fmt.Sprintf("/v2/%s/blobs/%s", w.ref.Context().RepositoryStr(), h.String()))

	resp, err := w.client.Head(u.String())
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusOK, http.StatusNotFound); err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// initiateUpload initiates the blob upload, which starts with a POST that can
// optionally include the hash of the layer and a list of repositories from
// which that layer might be read. On failure, an error is returned.
// On success, the layer was either mounted (nothing more to do) or a blob
// upload was initiated and the body of that blob should be sent to the returned
// location.
func (w *writer) initiateUpload(h v1.Hash) (location string, mounted bool, err error) {
	u := w.url(fmt.Sprintf("/v2/%s/blobs/uploads/", w.ref.Context().RepositoryStr()))
	uv := url.Values{
		"mount": []string{h.String()},
	}
	l, err := w.img.LayerByDigest(h)
	if err != nil {
		return "", false, err
	}

	if ml, ok := l.(*MountableLayer); ok {
		if w.ref.Context().RegistryStr() == ml.Reference.Context().RegistryStr() {
			uv["from"] = []string{ml.Reference.Context().RepositoryStr()}
		}
	}
	u.RawQuery = uv.Encode()

	// Make the request to initiate the blob upload.
	resp, err := w.client.Post(u.String(), "application/json", nil)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusCreated, http.StatusAccepted); err != nil {
		return "", false, err
	}

	// Check the response code to determine the result.
	switch resp.StatusCode {
	case http.StatusCreated:
		// We're done, we were able to fast-path.
		return "", true, nil
	case http.StatusAccepted:
		// Proceed to PATCH, upload has begun.
		loc, err := w.nextLocation(resp)
		return loc, false, err
	default:
		panic("Unreachable: initiateUpload")
	}
}

// streamBlob streams the contents of the blob to the specified location.
// On failure, this will return an error.  On success, this will return the location
// header indicating how to commit the streamed blob.
func (w *writer) streamBlob(h v1.Hash, streamLocation string) (commitLocation string, err error) {
	l, err := w.img.LayerByDigest(h)
	if err != nil {
		return "", err
	}
	blob, err := l.Compressed()
	if err != nil {
		return "", err
	}
	defer blob.Close()

	req, err := http.NewRequest(http.MethodPatch, streamLocation, blob)
	if err != nil {
		return "", err
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusNoContent, http.StatusAccepted, http.StatusCreated); err != nil {
		return "", err
	}

	// The blob has been uploaded, return the location header indicating
	// how to commit this layer.
	return w.nextLocation(resp)
}

// commitBlob commits this blob by sending a PUT to the location returned from streaming the blob.
func (w *writer) commitBlob(h v1.Hash, location string) (err error) {
	u, err := url.Parse(location)
	if err != nil {
		return err
	}
	v := u.Query()
	v.Set("digest", h.String())
	u.RawQuery = v.Encode()

	req, err := http.NewRequest(http.MethodPut, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return CheckError(resp, http.StatusCreated)
}

// uploadOne performs a complete upload of a single layer.
func (w *writer) uploadOne(h v1.Hash) error {
	existing, err := w.checkExisting(h)
	if err != nil {
		return err
	}
	if existing {
		log.Printf("existing blob: %v", h)
		return nil
	}

	location, mounted, err := w.initiateUpload(h)
	if err != nil {
		return err
	} else if mounted {
		log.Printf("mounted blob: %v", h)
		return nil
	}

	location, err = w.streamBlob(h, location)
	if err != nil {
		return err
	}

	if err := w.commitBlob(h, location); err != nil {
		return err
	}
	log.Printf("pushed blob %v", h)
	return nil
}

// commitImage does a PUT of the image's manifest.
func (w *writer) commitImage() error {
	raw, err := w.img.RawManifest()
	if err != nil {
		return err
	}
	mt, err := w.img.MediaType()
	if err != nil {
		return err
	}

	u := w.url(fmt.Sprintf("/v2/%s/manifests/%s", w.ref.Context().RepositoryStr(), w.ref.Identifier()))

	// Make the request to PUT the serialized manifest
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", string(mt))

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusOK, http.StatusCreated, http.StatusAccepted); err != nil {
		return err
	}

	digest, err := w.img.Digest()
	if err != nil {
		return err
	}

	// The image was successfully pushed!
	log.Printf("%v: digest: %v size: %d", w.ref, digest, len(raw))
	return nil
}

func scopesForUploadingImage(ref name.Reference, layers []v1.Layer) []string {
	// use a map as set to remove duplicates scope strings
	scopeSet := map[string]struct{}{}

	for _, l := range layers {
		if ml, ok := l.(*MountableLayer); ok {
			// we add push scope for ref.Context() after the loop
			if ml.Reference.Context() != ref.Context() {
				scopeSet[ml.Reference.Context().Scope(transport.PullScope)] = struct{}{}
			}
		}
	}

	scopes := make([]string, 0)
	// Push scope should be the first element because a few registries just look at the first scope to determine access.
	scopes = append(scopes, ref.Scope(transport.PushScope))

	for scope, _ := range scopeSet {
		scopes = append(scopes, scope)
	}

	return scopes
}

// TODO(mattmoor): WriteIndex
