/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

// This file implements a go-getter wrapper for cheggaaa progress bar
// based on:
// https://github.com/hashicorp/go-getter/blob/master/cmd/go-getter/progress_tracking.go

package download

import (
	"io"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb/v3"
	"github.com/hashicorp/go-getter"
)

// DefaultProgressBar is the default cheggaaa progress bar
var DefaultProgressBar getter.ProgressTracker = &progressBar{}

type progressBar struct {
	lock     sync.Mutex
	progress *pb.ProgressBar
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.
func (cpb *progressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()
	if cpb.progress == nil {
		cpb.progress = pb.New64(totalSize)
	}
	p := pb.Full.Start64(totalSize)
	fn := filepath.Base(src)
	// abbreviate filename for progress
	maxwidth := 30 - len("...")
	if len(fn) > maxwidth {
		fn = fn[0:maxwidth] + "..."
	}
	p.Set("prefix", "    > "+fn+": ")
	p.SetCurrent(currentSize)
	p.Set(pb.Bytes, true)

	// Just a hair less than 80 (standard terminal width) for aesthetics & pasting into docs
	p.SetWidth(79)
	barReader := p.NewProxyReader(stream)

	return &readCloser{
		Reader: barReader,
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()
			p.Finish()
			return nil
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (c *readCloser) Close() error { return c.close() }
