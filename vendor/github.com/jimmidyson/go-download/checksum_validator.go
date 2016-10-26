//    Copyright 2016 Red Hat, Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package download

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type checksumValidator interface {
	io.Writer
	validate() bool
}

func newValidator(hasher hash.Hash, client *http.Client, checksum, filename string) (checksumValidator, error) {
	if u, err := url.Parse(checksum); err == nil && len(u.Scheme) != 0 {
		if u.Scheme == "http" || u.Scheme == "https" {
			return newValidatorFromChecksumURL(hasher, client, checksum, filename)
		}

		return nil, errors.Errorf("unsupported scheme: %s (supported schemes: %v)", u.Scheme, []string{"http", "https"})
	}

	if _, err := hex.DecodeString(checksum); err == nil {
		return &validator{
			hasher:   hasher,
			checksum: checksum,
		}, nil
	}

	if f, err := os.Open(checksum); err == nil {
		defer func() { _ = f.Close() }() // #nosec
		return newValidatorFromReader(hasher, f, filename)
	}

	return nil, errors.New("invalid checksum: must be one of hex encoded checksum, URL or file path")
}

func newValidatorFromChecksumURL(hasher hash.Hash, client *http.Client, checksumURL, filename string) (checksumValidator, error) {
	resp, err := client.Get(checksumURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download checksum file")
	}
	defer func() { _ = resp.Body.Close() }() // #nosec

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to download checksum file: received status code %d", resp.StatusCode)
	}

	return newValidatorFromReader(hasher, resp.Body, filename)
}

func newValidatorFromReader(hasher hash.Hash, reader io.Reader, filename string) (checksumValidator, error) {
	scanner := bufio.NewScanner(reader)
	var b bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		spl := strings.Fields(line)
		if len(spl) == 2 {
			if spl[1] == filename {
				trimmedHash := strings.TrimSpace(spl[0])
				if _, err := hex.DecodeString(trimmedHash); err == nil {
					return &validator{
						hasher:   hasher,
						checksum: trimmedHash,
					}, nil
				}
			}
		}
		if b.Len() == 0 {
			_, _ = b.WriteString(line) // #nosec
		}
	}
	buf := b.String()
	if len(buf) > 0 {
		trimmedHash := strings.TrimSpace(buf)
		if _, err := hex.DecodeString(trimmedHash); err == nil {
			return &validator{
				hasher:   hasher,
				checksum: trimmedHash,
			}, nil
		}
	}

	return nil, errors.New("failed to retrieve checksum")
}

var _ checksumValidator = &validator{}

type validator struct {
	hasher   hash.Hash
	checksum string
}

func (v *validator) validate() bool {
	return hex.EncodeToString(v.hasher.Sum(nil)) == v.checksum
}

func (v *validator) Write(p []byte) (n int, err error) {
	return v.hasher.Write(p)
}
