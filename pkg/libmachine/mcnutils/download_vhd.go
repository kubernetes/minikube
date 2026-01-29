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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s.io/minikube/pkg/libmachine/log"
)

type ProgressWriter struct {
	Total      int64
	Downloaded int64
	mu         sync.Mutex
	TargetName string
}

func NewProgressWriter(total int64, targetName string) *ProgressWriter {
	return &ProgressWriter{Total: total, TargetName: targetName}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.mu.Lock()
	pw.Downloaded += int64(n)
	pw.printProgress()
	pw.mu.Unlock()
	return n, nil
}

func (pw *ProgressWriter) printProgress() {
	// Overwrite the same line with \r
	fmt.Printf("\r        > %s: %d / %d bytes complete", pw.TargetName, pw.Downloaded, pw.Total)
}

// copyLocalFile copies from a local source path to destination, reporting progress.
func copyLocalFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open local source file %q: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Get total size
	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local source file %q: %w", srcPath, err)
	}
	totalSize := info.Size()

	outFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", dstPath, err)
	}
	defer outFile.Close()

	pw := NewProgressWriter(totalSize, filepath.Base(dstPath))
	// Use TeeReader: read from srcFile, write to pw (for progress), and to outFile
	_, err = io.Copy(outFile, io.TeeReader(srcFile, pw))
	if err != nil {
		return fmt.Errorf("error copying local file to %q: %w", dstPath, err)
	}

	// Final newline after progress
	fmt.Printf("\n")
	log.Infof("\t> Local copy complete: %s\n", dstPath)
	return nil
}

// DownloadPart downloads a byte-range [start,end] of the URL into a temporary part file.
// On error, it returns the error; progress for the range is reported via pw.
// The caller goroutine must call wg.Done() exactly once.
func DownloadPart(urlStr string, start, end int64, partFileName string, pw *ProgressWriter, retryLimit int) error {
	var resp *http.Response
	var err error

	// Retry loop for downloading a part
	for retries := 0; retries <= retryLimit; retries++ {
		req, errReq := http.NewRequest("GET", urlStr, nil)
		if errReq != nil {
			log.Errorf("Error creating request: %v", errReq)
			return errReq
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("Error downloading part (attempt %d): %v", retries+1, err)
			// Exponential backoff before retry
			time.Sleep(time.Duration(1<<uint(retries)) * time.Second)
			continue
		}

		// If we get Partial Content, proceed to save
		if resp.StatusCode == http.StatusPartialContent {
			partFile, errCreate := os.Create(partFileName)
			if errCreate != nil {
				log.Errorf("Error creating part file: %v", errCreate)
				resp.Body.Close()
				return errCreate
			}
			// Copy with progress
			buf := make([]byte, 32*1024) // 32 KB buffer
			_, errCopy := io.CopyBuffer(io.MultiWriter(partFile, pw), resp.Body, buf)
			partFile.Close()
			resp.Body.Close()
			if errCopy != nil {
				log.Errorf("Error saving part: %v", errCopy)
				return errCopy
			}
			return nil // success
		}

		// Unexpected status => retry
		log.Errorf("Error: expected status 206 Partial Content, got %d", resp.StatusCode)
		resp.Body.Close()
		time.Sleep(time.Duration(1<<uint(retries)) * time.Second)
	}

	log.Errorf("Failed to download part after %d retries", retryLimit)
	return fmt.Errorf("failed to download part after %d retries", retryLimit)
}

// DownloadVHDX downloads a VHD from a URL or copies from a local path if detected.
//   - urlStr: can be HTTP(S) URL or a local filesystem path (absolute or relative), or file:// URI.
//   - filePath: destination path to write the VHD.
//   - numParts: for HTTP downloads, number of parallel parts; ignored for local copy.
//   - retryLimit: retry count per part.
//
// It returns an error on failure.
func DownloadVHDX(urlStr string, filePath string, numParts int, retryLimit int) error {
	// First, check if urlStr is a local file path or file:// URI.
	if u, err := url.Parse(urlStr); err == nil {
		if u.Scheme == "file" {
			// file:// URI: extract path
			localPath := u.Path
			// On Windows, file:///C:/path yields u.Path="/C:/path". Strip leading slash.
			if strings.HasPrefix(localPath, "/") && os.PathSeparator == '\\' && len(localPath) > 2 && localPath[1] == ':' {
				localPath = localPath[1:]
			}
			return copyLocalFile(localPath, filePath)
		}
	}
	// If no scheme or non-file scheme, check if it's a path existing on disk:
	if fi, err := os.Stat(urlStr); err == nil && !fi.IsDir() {
		// Treat as local file
		return copyLocalFile(urlStr, filePath)
	}

	// Otherwise assume HTTP(S) URL:
	// First, HEAD to get total size
	resp, err := http.Head(urlStr)
	if err != nil {
		return fmt.Errorf("failed to get file info from URL %q: %w", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status from HEAD %q: %s", urlStr, resp.Status)
	}
	totalSize := resp.ContentLength
	if totalSize <= 0 {
		return fmt.Errorf("unknown content length for URL %q", urlStr)
	}

	// For progress display: use base name of destination
	pw := NewProgressWriter(totalSize, filepath.Base(filePath))

	// Partition download into numParts
	partSize := totalSize / int64(numParts)
	var wg sync.WaitGroup
	var muErr sync.Mutex
	downloadErrors := make([]error, 0, numParts)
	partFiles := make([]string, numParts)

	for i := 0; i < numParts; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if i == numParts-1 {
			end = totalSize - 1
		}
		partFileName := fmt.Sprintf("%s.part-%d.tmp", filePath, i)
		partFiles[i] = partFileName

		wg.Add(1)
		go func(idx int, s, e int64, pfn string) {
			defer wg.Done()
			errPart := DownloadPart(urlStr, s, e, pfn, pw, retryLimit)
			if errPart != nil {
				muErr.Lock()
				downloadErrors = append(downloadErrors, fmt.Errorf("part %d: %w", idx, errPart))
				muErr.Unlock()
			}
		}(i, start, end, partFileName)
	}

	// Wait for all parts
	wg.Wait()

	if len(downloadErrors) > 0 {
		// Clean up partial files
		for _, pfn := range partFiles {
			os.Remove(pfn)
		}
		return fmt.Errorf("download failed for parts: %v", downloadErrors)
	}

	// Merge parts
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", filePath, err)
	}
	defer outFile.Close()

	for _, pfn := range partFiles {
		f, errOpen := os.Open(pfn)
		if errOpen != nil {
			return fmt.Errorf("failed to open part file %q: %w", pfn, errOpen)
		}
		_, errCopy := io.Copy(outFile, f)
		f.Close()
		if errCopy != nil {
			return fmt.Errorf("failed to merge part file %q: %w", pfn, errCopy)
		}
		os.Remove(pfn)
	}

	// Final newline after progress
	fmt.Printf("\n")
	log.Infof("\t> Download complete: %s\n", filePath)
	return nil
}
