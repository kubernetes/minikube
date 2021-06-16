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

	"github.com/cheggaaa/pb"
	"github.com/hashicorp/go-getter"
	"k8s.io/klog/v2"
)

// DefaultProgressBar is the default cheggaaa progress bar
var DefaultProgressBar getter.ProgressTracker = &progressBar{}

type progressBar struct {
	lock sync.Mutex
	pool *pb.Pool
	pbs  int
}

// AddProgressBar add progress bar to the concurrent pool
func (cpb *progressBar) AddProgressBar(p *pb.ProgressBar) error {
	if cpb.pool == nil {
		cpb.pool = pb.NewPool()
		if err := cpb.pool.Start(); err != nil {
			return err
		}
	}
	cpb.pool.Add(p)
	cpb.pbs++
	return nil
}

// RemoveProgressBar removes progress bar from the concurrent pool
func (cpb *progressBar) RemoveProgressBar(p *pb.ProgressBar) error {
	cpb.pbs--
	if cpb.pbs <= 0 {
		if err := cpb.pool.Stop(); err != nil {
			return err
		}
		cpb.pool = nil
	}
	return nil
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.
func (cpb *progressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()
	//p := pb.Full.Start64(totalSize)
	p := pb.New64(totalSize)
	p.ShowSpeed = true
	p.ShowTimeLeft = true
	fn := filepath.Base(src)
	// abbreviate filename for progress
	maxwidth := 30 - len("...")
	if len(fn) > maxwidth {
		fn = fn[0:maxwidth] + "..."
	}
	p.Prefix("    > " + fn + ": ")
	p.Set64(currentSize)
	p.SetUnits(pb.U_BYTES)

	// Just a hair less than 80 (standard terminal width) for aesthetics & pasting into docs
	p.SetWidth(79)
	p.Start()
	barReader := p.NewProxyReader(stream)

	if err := cpb.AddProgressBar(p); err != nil {
		klog.Errorf("pool start: %v", err)
	}
	return &readCloser{
		Reader: barReader,
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()
			p.Finish()
			if err := cpb.RemoveProgressBar(p); err != nil {
				klog.Errorf("pool stop: %v", err)
			}
			return nil
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (c *readCloser) Close() error { return c.close() }
