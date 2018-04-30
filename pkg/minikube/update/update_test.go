/*
Copyright 2018 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package update

import (
	"fmt"
	"io/ioutil"
	pkgtesting "k8s.io/minikube/pkg/testing"
	"k8s.io/minikube/pkg/util"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	testVersion = "v0.24.0"
)

var (
	_, b, _, _         = runtime.Caller(0)
	basepath           = filepath.Dir(b)
	err                error
	testDir            string
	expectedBinaryPath string
	testData           = []struct {
		platform   string
		version    string
		binaryName string
	}{
		{"linux", testVersion, "minikube-linux-amd64"},
		{"darwin", testVersion, "minikube-darwin-amd64"},
		// TODO: Add windows once Issue #2536 is fixed
	}
)

func TestDownloadBinary(t *testing.T) {
	mockTransport := pkgtesting.NewMockRoundTripper()
	addMockResponses(mockTransport)
	client := http.DefaultClient
	client.Transport = mockTransport
	defer pkgtesting.ResetDefaultRoundTripper()

	setUp(t)
	defer os.RemoveAll(testDir)

	for _, tt := range testData {
		expectedBinaryPath = filepath.Join(testDir, tt.binaryName)
		downloadURL := util.GetBinaryDownloadURL(tt.version, tt.platform)
		binaryPath, err := downloadBinary(downloadURL, testDir)
		if err != nil {
			t.Fatalf("failed to download binary: %s", err)
		}
		if expectedBinaryPath != binaryPath {
			t.Fatalf("expected binary path %s but got %s", expectedBinaryPath, binaryPath)
		}
	}
}

func setUp(t *testing.T) {
	testDir, err = ioutil.TempDir("", "minishift-test-")
	if err != nil {
		t.Fatal(err)
	}
}

func addMockResponses(mockTransport *pkgtesting.MockRoundTripper) {
	binaryData := []struct {
		version  string
		platform string
		content  string
		checksum string
	}{
		{testVersion, "linux", "minikube-linux-amd64", "de576a9dca4a9870299106529d9712b94ce424380205ef2cb7074bdc46d1b549"},
		{testVersion, "darwin", "minikube-darwin-amd64", "9c13083bf4c949f0f4849c25e1acee33a3b091faad83735a54492923a630ff36"},
		// TODO: Add windows once Issue #2536 is fixed
	}

	for _, binary := range binaryData {
		binaryUrl := util.GetBinaryDownloadURL(binary.version, binary.platform)
		mockTransport.RegisterResponse(binaryUrl, &pkgtesting.CannedResponse{
			ResponseType: pkgtesting.ServeString,
			Response:     binary.content,
			ContentType:  pkgtesting.OctetStream,
		})
		shaUrl := fmt.Sprintf("%s.sha256", binaryUrl)
		mockTransport.RegisterResponse(shaUrl, &pkgtesting.CannedResponse{
			ResponseType: pkgtesting.ServeString,
			Response:     binary.checksum,
			ContentType:  pkgtesting.OctetStream,
		})
	}
}
