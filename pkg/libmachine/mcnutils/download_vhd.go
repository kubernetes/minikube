/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package mcnutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
	"k8s.io/minikube/pkg/libmachine/log"
)

const (
	vhdTerminalInterval = 500 * time.Millisecond
	vhdLogInterval      = 5 * time.Second
	vhdCopyBufferSize   = 32 * 1024
)

type ProgressWriter struct {
	Total      int64
	Downloaded int64
	TargetName string

	mu          sync.Mutex
	output      io.Writer
	interactive bool
	now         func() time.Time
	interval    time.Duration
	lastUpdate  time.Time
	action      string
	finished    bool
	logError    func(error)
}

func NewProgressWriter(total int64, targetName string) *ProgressWriter {
	return newVHDProgress(total, targetName, os.Stdout, isatty.IsTerminal(os.Stdout.Fd()), time.Now)
}

func newVHDProgress(total int64, targetName string, output io.Writer, interactive bool, now func() time.Time) *ProgressWriter {
	interval := vhdLogInterval
	if interactive {
		interval = vhdTerminalInterval
	}
	return &ProgressWriter{
		Total: total, TargetName: filepath.Base(targetName),
		output: output, interactive: interactive, now: now,
		interval: interval, lastUpdate: now(),
		logError: func(err error) { log.Errorf("Error writing VHD progress: %v", err) },
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.Downloaded += int64(len(p))
	if pw.output == nil || pw.finished {
		return len(p), nil
	}
	now := pw.now()
	if now.Sub(pw.lastUpdate) >= pw.interval {
		pw.lastUpdate = now
		prefix, suffix := "", "\n"
		if pw.interactive {
			prefix, suffix = "\r", ""
		}
		pw.print(fmt.Sprintf("%s    > %s: %d / %d bytes transferred%s", prefix, pw.TargetName, pw.Downloaded, pw.Total, suffix))
	}
	return len(p), nil
}

// print is called with mu held. A broken UI must not abort an otherwise valid transfer.
func (pw *ProgressWriter) print(text string) {
	if pw.output == nil {
		return
	}
	n, err := io.WriteString(pw.output, text)
	if err == nil && n != len(text) {
		err = io.ErrShortWrite
	}
	if err != nil {
		pw.output = nil
		pw.logError(err)
	}
}

func (pw *ProgressWriter) start(action string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.action = action
	pw.print(fmt.Sprintf("%s started: %s (%d bytes)\n", action, pw.TargetName, pw.Total))
}

func (pw *ProgressWriter) finish(err error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	if pw.finished {
		return
	}
	pw.finished = true
	separator := ""
	if pw.interactive {
		separator = "\n"
	}
	outcome := "complete"
	if err != nil {
		outcome = "failed"
	}
	pw.print(fmt.Sprintf("%s%s %s: %s (%d / %d bytes transferred)\n", separator, pw.action, outcome, pw.TargetName, pw.Downloaded, pw.Total))
}

type vhdProgressDestination struct {
	dst io.Writer
	pw  *ProgressWriter
}

func (w vhdProgressDestination) Write(p []byte) (int, error) {
	n, err := w.dst.Write(p)
	_, _ = w.pw.Write(p[:n])
	if n != len(p) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

// writeVHD publishes only a closed, successful acquisition. In particular, a failed
// download must not leave a file that the existence-based VHD cache will reuse.
func writeVHD(dstPath string, pw *ProgressWriter, action string, write func(*os.File) error) (err error) {
	f, err := os.CreateTemp(filepath.Dir(dstPath), filepath.Base(dstPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary VHD: %w", err)
	}
	closed, published := false, false
	if pw != nil {
		pw.start(action)
		defer func() { pw.finish(err) }()
	}
	defer func() {
		if !closed {
			if closeErr := f.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close temporary VHD: %w", closeErr))
			}
		}
		if !published {
			if removeErr := os.Remove(f.Name()); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("remove temporary VHD: %w", removeErr))
			}
		}
	}()
	if err := write(f); err != nil {
		return err
	}
	err = f.Close()
	closed = true
	if err != nil {
		return fmt.Errorf("close temporary VHD: %w", err)
	}
	if err := os.Rename(f.Name(), dstPath); err != nil {
		return fmt.Errorf("publish VHD %q: %w", dstPath, err)
	}
	published = true
	return nil
}

// copyLocalFile copies from a local source path to destination, reporting progress.
func copyLocalFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open local VHD %q: %w", srcPath, err)
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat local VHD %q: %w", srcPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("local VHD %q is not a regular file", srcPath)
	}
	if dstInfo, err := os.Stat(dstPath); err == nil && os.SameFile(info, dstInfo) {
		return fmt.Errorf("local VHD source and destination are the same file: %q", srcPath)
	}
	pw := NewProgressWriter(info.Size(), filepath.Base(dstPath))
	return writeVHD(dstPath, pw, "Copy", func(dst *os.File) error {
		n, err := io.CopyBuffer(vhdProgressDestination{dst: dst, pw: pw}, src, make([]byte, vhdCopyBufferSize))
		if err != nil {
			return fmt.Errorf("copy local VHD: %w", err)
		}
		if n != info.Size() {
			return fmt.Errorf("local VHD size changed: copied %d bytes, expected %d", n, info.Size())
		}
		return nil
	})
}

// DownloadPart downloads an inclusive byte range into a part file. The caller owns
// the shared progress lifecycle; failed acquisitions do not publish the part file.
func DownloadPart(urlStr string, start, end int64, partFileName string, pw *ProgressWriter, retryLimit int) error {
	if pw == nil || start < 0 || end < start || end >= pw.Total || retryLimit < 0 {
		return fmt.Errorf("invalid VHD range or retry limit")
	}
	return writeVHD(partFileName, nil, "", func(dst *os.File) error {
		return downloadVHDRange(context.Background(), urlStr, start, end, pw.Total, "", dst, pw, retryLimit)
	})
}

func downloadVHDRange(ctx context.Context, urlStr string, start, end, total int64, validator string, dst io.Writer, pw *ProgressWriter, retryLimit int) error {
	backoff := time.Second
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return fmt.Errorf("create VHD range request: %w", err)
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
		req.Header.Set("Accept-Encoding", "identity")
		if validator != "" {
			req.Header.Set("If-Range", validator)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusPartialContent {
			defer resp.Body.Close()
			return copyVHDRange(resp, start, end, total, dst, pw)
		}
		retryable := err != nil
		if err == nil {
			err = fmt.Errorf("expected HTTP 206 Partial Content, got %s", resp.Status)
			retryable = resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode < 600)
			resp.Body.Close()
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !retryable || attempt >= retryLimit {
			return fmt.Errorf("VHD range %d-%d failed after %d attempt(s): %w", start, end, attempt+1, err)
		}
		pw.mu.Lock()
		log.Errorf("Error downloading VHD range %d-%d (attempt %d); retrying: %v", start, end, attempt+1, err)
		pw.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, 30*time.Second)
	}
}

func copyVHDRange(resp *http.Response, start, end, total int64, dst io.Writer, pw *ProgressWriter) error {
	wantRange := fmt.Sprintf("bytes %d-%d/%d", start, end, total)
	if got := resp.Header.Get("Content-Range"); got != wantRange {
		return fmt.Errorf("invalid VHD Content-Range %q; expected %q", got, wantRange)
	}
	if encoding := resp.Header.Get("Content-Encoding"); encoding != "" && encoding != "identity" {
		return fmt.Errorf("unexpected VHD Content-Encoding %q", encoding)
	}
	size := end - start + 1
	if resp.ContentLength >= 0 && resp.ContentLength != size {
		return fmt.Errorf("invalid VHD range length %d; expected %d", resp.ContentLength, size)
	}
	// Bound each writer to its assigned region even if a server sends an oversized body.
	n, err := io.CopyBuffer(vhdProgressDestination{dst: dst, pw: pw}, io.LimitReader(resp.Body, size), make([]byte, vhdCopyBufferSize))
	if err != nil {
		return fmt.Errorf("copy VHD range %d-%d: %w", start, end, err)
	}
	if n != size {
		return fmt.Errorf("short VHD range %d-%d: received %d bytes, expected %d", start, end, n, size)
	}
	var extra [1]byte
	if _, err := io.ReadFull(resp.Body, extra[:]); err != io.EOF {
		if err != nil {
			return fmt.Errorf("read VHD range end: %w", err)
		}
		return fmt.Errorf("VHD range %d-%d exceeds expected length", start, end)
	}
	return nil
}

// DownloadVHDX acquires a VHD from HTTP(S), a local path, or a file:// URI.
// HTTP ranges are written concurrently into a single temporary destination.
// numParts and retryLimit apply only to HTTP acquisitions.
func DownloadVHDX(urlStr, filePath string, numParts, retryLimit int) error {
	if _, err := os.Stat(urlStr); err == nil || filepath.IsAbs(urlStr) {
		return copyLocalFile(urlStr, filePath)
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("parse VHD source: %w", err)
	}
	switch u.Scheme {
	case "":
		return copyLocalFile(urlStr, filePath)
	case "file":
		localPath := u.Path
		if os.PathSeparator == '\\' && len(localPath) > 2 && localPath[0] == '/' && localPath[2] == ':' {
			localPath = localPath[1:]
		}
		if u.Host != "" && u.Host != "localhost" {
			if os.PathSeparator != '\\' {
				return fmt.Errorf("remote file URI hosts are not supported: %q", u.Host)
			}
			localPath = "//" + u.Host + u.Path
		}
		return copyLocalFile(filepath.FromSlash(localPath), filePath)
	case "http", "https":
		return downloadHTTPVHD(urlStr, filePath, numParts, retryLimit)
	default:
		return fmt.Errorf("unsupported VHD source scheme %q", u.Scheme)
	}
}

func downloadHTTPVHD(urlStr, filePath string, numParts, retryLimit int) error {
	if numParts <= 0 || retryLimit < 0 {
		return fmt.Errorf("VHD part count must be positive and retry limit must be nonnegative")
	}
	req, err := http.NewRequest(http.MethodHead, urlStr, nil)
	if err != nil {
		return fmt.Errorf("create VHD metadata request: %w", err)
	}
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("get VHD metadata: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status from VHD HEAD: %s", resp.Status)
	}
	total := resp.ContentLength
	if total <= 0 {
		return fmt.Errorf("unknown or empty VHD content length: %d", total)
	}
	validator := resp.Header.Get("ETag")
	if validator == "" || strings.HasPrefix(validator, "W/") {
		validator = resp.Header.Get("Last-Modified")
	}
	if int64(numParts) > total {
		numParts = int(total)
	}
	pw := NewProgressWriter(total, filepath.Base(filePath))
	return writeVHD(filePath, pw, "Download", func(dst *os.File) error {
		if err := dst.Truncate(total); err != nil {
			return fmt.Errorf("size temporary VHD: %w", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var wg sync.WaitGroup
		var firstError sync.Once
		var downloadErr error
		partSize := total / int64(numParts)
		for i := 0; i < numParts; i++ {
			start := int64(i) * partSize
			end := start + partSize - 1
			if i == numParts-1 {
				end = total - 1
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := downloadVHDRange(ctx, urlStr, start, end, total, validator, io.NewOffsetWriter(dst, start), pw, retryLimit)
				if err != nil {
					firstError.Do(func() {
						downloadErr = fmt.Errorf("part %d: %w", i, err)
						cancel()
					})
				}
			}()
		}
		wg.Wait()
		return downloadErr
	})
}
