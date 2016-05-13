package docker2aci

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"

	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

const (
	hashPrefix = "sha512-"
)

type aciInfo struct {
	path          string
	key           string
	ImageManifest *schema.ImageManifest
}

// ConversionStore is an simple implementation of the acirenderer.ACIRegistry
// interface. It stores the Docker layers converted to ACI so we can take
// advantage of acirenderer to generate a squashed ACI Image.
type ConversionStore struct {
	acis map[string]*aciInfo
}

func NewConversionStore() *ConversionStore {
	return &ConversionStore{acis: make(map[string]*aciInfo)}
}

func (ms *ConversionStore) WriteACI(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	cr, err := aci.NewCompressedReader(f)
	if err != nil {
		return "", err
	}
	defer cr.Close()

	h := sha512.New()
	r := io.TeeReader(cr, h)

	// read the file so we can get the hash
	if _, err := io.Copy(ioutil.Discard, r); err != nil {
		return "", fmt.Errorf("error reading ACI: %v", err)
	}

	im, err := aci.ManifestFromImage(f)
	if err != nil {
		return "", err
	}

	key := ms.HashToKey(h)
	ms.acis[key] = &aciInfo{path: path, key: key, ImageManifest: im}
	return key, nil
}

func (ms *ConversionStore) GetImageManifest(key string) (*schema.ImageManifest, error) {
	aci, ok := ms.acis[key]
	if !ok {
		return nil, fmt.Errorf("aci with key: %s not found", key)
	}
	return aci.ImageManifest, nil
}

func (ms *ConversionStore) GetACI(name types.ACIdentifier, labels types.Labels) (string, error) {
	for _, aci := range ms.acis {
		// we implement this function to comply with the interface so don't
		// bother implementing a proper label check
		if aci.ImageManifest.Name.String() == name.String() {
			return aci.key, nil
		}
	}
	return "", fmt.Errorf("aci not found")
}

func (ms *ConversionStore) ReadStream(key string) (io.ReadCloser, error) {
	img, ok := ms.acis[key]
	if !ok {
		return nil, fmt.Errorf("stream for key: %s not found", key)
	}
	f, err := os.Open(img.path)
	if err != nil {
		return nil, fmt.Errorf("error opening aci: %s", img.path)
	}

	tr, err := aci.NewCompressedReader(f)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

func (ms *ConversionStore) ResolveKey(key string) (string, error) {
	return key, nil
}

func (ms *ConversionStore) HashToKey(h hash.Hash) string {
	s := h.Sum(nil)
	return fmt.Sprintf("%s%x", hashPrefix, s)
}
