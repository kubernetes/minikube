// This file implements a go-getter wrapper for cheggaaa progress bar

// based on:
// https://github.com/hashicorp/go-getter/blob/master/cmd/go-getter/progress_tracking.go
package util

import (
	"io"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
)

var defaultProgressBar getter.ProgressTracker = &progressBar{}

type progressBar struct {
	lock sync.Mutex
	pool *pb.Pool
	pbs  int
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.
func (cpb *progressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()

	if cpb.pool == nil {
		cpb.pool = pb.NewPool()
		if err := cpb.pool.Start(); err != nil {
			glog.Errorf("pool start: %v", err)
		}
	}

	p := pb.New64(totalSize)
	p.Set64(currentSize)
	p.SetUnits(pb.U_BYTES)
	p.Prefix(filepath.Base(src))

	cpb.pool.Add(p)
	reader := p.NewProxyReader(stream)

	cpb.pbs++
	return &readCloser{
		Reader: reader,
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()

			p.Finish()
			cpb.pbs--
			if cpb.pbs <= 0 {
				if err := cpb.pool.Stop(); err != nil {
					glog.Errorf("pool stop: %v", err)
				}
				cpb.pool = nil
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
