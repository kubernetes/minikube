/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package download

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/go-getter"
	"k8s.io/minikube/pkg/minikube/out/register"
)

var DefaultJSONOutput getter.ProgressTracker = &jsonOutput{}

type jsonOutput struct {
	lock sync.Mutex
}

// TrackProgress prints out progress of the stream in JSON format until closed
func (cpb *jsonOutput) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()

	register.PrintDownload(src)

	return &readCloser{
		Reader: &jsonReader{
			Reader:   stream,
			artifact: src,
			current:  currentSize,
			total:    totalSize,
			Time:     time.Now(),
		},
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()
			return nil
		},
	}
}

// jsonReader is a wrapper for printing with JSON output
type jsonReader struct {
	artifact string
	current  int64
	total    int64
	io.Reader
	time.Time
}

func (r *jsonReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.current += int64(n)
	progress := float64(r.current) / float64(r.total)
	// print progress every second so user isn't overwhelmed with events
	if t := time.Now(); t.Sub(r.Time) > time.Second || progress == 1 {
		register.PrintDownloadProgress(r.artifact, fmt.Sprintf("%v", progress))
		r.Time = t
	}
	return
}
