// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
)

type MockProgressTracking struct {
	sync.Mutex
	downloaded map[string]int
}

func (p *MockProgressTracking) TrackProgress(src string,
	currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	p.Lock()
	defer p.Unlock()

	if p.downloaded == nil {
		p.downloaded = map[string]int{}
	}

	v := p.downloaded[src]
	p.downloaded[src] = v + 1
	return stream
}

func TestGet_progress(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// all good
		rw.Header().Add("X-Terraform-Get", "something")
	}))
	defer s.Close()

	{ // dl without tracking
		dst := filepath.Join(t.TempDir(), "test-file")
		if err := GetFile(dst, s.URL+"/file?thig=this&that"); err != nil {
			t.Fatalf("download failed: %v", err)
		}
	}

	{ // tracking
		p := &MockProgressTracking{}
		dst := filepath.Join(t.TempDir(), "test-file")
		if err := GetFile(dst, s.URL+"/file?thig=this&that", WithProgress(p)); err != nil {
			t.Fatalf("download failed: %v", err)
		}
		if err := GetFile(dst, s.URL+"/otherfile?thig=this&that", WithProgress(p)); err != nil {
			t.Fatalf("download failed: %v", err)
		}

		if p.downloaded["file"] != 1 {
			t.Error("Expected a file download")
		}
		if p.downloaded["otherfile"] != 1 {
			t.Error("Expected a otherfile download")
		}
	}
}
