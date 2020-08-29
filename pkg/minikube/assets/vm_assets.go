/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package assets

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// MemorySource is the source name used for in-memory copies
const MemorySource = "memory"

// CopyableFile is something that can be copied
type CopyableFile interface {
	io.Reader
	GetLength() int
	GetSourcePath() string

	GetTargetDir() string
	GetTargetName() string
	GetPermissions() string
	GetModTime() (time.Time, error)
	Seek(int64, int) (int64, error)
}

// BaseAsset is the base asset class
type BaseAsset struct {
	SourcePath  string
	TargetDir   string
	TargetName  string
	Permissions string
	Source      string
}

// GetSourcePath returns asset name
func (b *BaseAsset) GetSourcePath() string {
	return b.SourcePath
}

// GetTargetDir returns target dir
func (b *BaseAsset) GetTargetDir() string {
	return b.TargetDir
}

// GetTargetName returns target name
func (b *BaseAsset) GetTargetName() string {
	return b.TargetName
}

// GetPermissions returns permissions
func (b *BaseAsset) GetPermissions() string {
	return b.Permissions
}

// GetModTime returns mod time
func (b *BaseAsset) GetModTime() (time.Time, error) {
	return time.Time{}, nil
}

// FileAsset is an asset using a file
type FileAsset struct {
	BaseAsset
	reader io.ReadSeeker
}

// NewMemoryAssetTarget creates a new MemoryAsset, with target
func NewMemoryAssetTarget(d []byte, targetPath, permissions string) *MemoryAsset {
	return NewMemoryAsset(d, path.Dir(targetPath), path.Base(targetPath), permissions)
}

// NewFileAsset creates a new FileAsset
func NewFileAsset(src, targetDir, targetName, permissions string) (*FileAsset, error) {
	glog.V(4).Infof("NewFileAsset: %s -> %s", src, path.Join(targetDir, targetName))

	f, err := os.Open(src)
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}

	info, err := os.Stat(src)
	if err != nil {
		return nil, errors.Wrapf(err, "stat")
	}

	if info.Size() == 0 {
		glog.Warningf("NewFileAsset: %s is an empty file!", src)
	}

	return &FileAsset{
		BaseAsset: BaseAsset{
			SourcePath:  src,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
		reader: io.NewSectionReader(f, 0, info.Size()),
	}, nil
}

// GetLength returns the file length, or 0 (on error)
func (f *FileAsset) GetLength() (flen int) {
	fi, err := os.Stat(f.SourcePath)
	if err != nil {
		glog.Errorf("stat(%q) failed: %v", f.SourcePath, err)
		return 0
	}
	return int(fi.Size())
}

// GetModTime returns modification time of the file
func (f *FileAsset) GetModTime() (time.Time, error) {
	fi, err := os.Stat(f.SourcePath)
	if err != nil {
		glog.Errorf("stat(%q) failed: %v", f.SourcePath, err)
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// Read reads the asset
func (f *FileAsset) Read(p []byte) (int, error) {
	if f.reader == nil {
		return 0, errors.New("Error attempting FileAsset.Read, FileAsset.reader uninitialized")
	}
	return f.reader.Read(p)
}

// Seek resets the reader to offset
func (f *FileAsset) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

// MemoryAsset is a memory-based asset
type MemoryAsset struct {
	BaseAsset
	reader io.ReadSeeker
	length int
}

// GetLength returns length
func (m *MemoryAsset) GetLength() int {
	return m.length
}

// Read reads the asset
func (m *MemoryAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

// Seek resets the reader to offset
func (m *MemoryAsset) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

// NewMemoryAsset creates a new MemoryAsset
func NewMemoryAsset(d []byte, targetDir, targetName, permissions string) *MemoryAsset {
	return &MemoryAsset{
		BaseAsset: BaseAsset{
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
			SourcePath:  MemorySource,
		},
		reader: bytes.NewReader(d),
		length: len(d),
	}
}

// BinAsset is a bindata (binary data) asset
type BinAsset struct {
	BaseAsset
	reader   io.ReadSeeker
	template *template.Template
	length   int
}

// MustBinAsset creates a new BinAsset, or panics if invalid
func MustBinAsset(name, targetDir, targetName, permissions string, isTemplate bool) *BinAsset {
	asset, err := NewBinAsset(name, targetDir, targetName, permissions, isTemplate)
	if err != nil {
		panic(fmt.Sprintf("Failed to define asset %s: %v", name, err))
	}
	return asset
}

// NewBinAsset creates a new BinAsset
func NewBinAsset(name, targetDir, targetName, permissions string, isTemplate bool) (*BinAsset, error) {
	m := &BinAsset{
		BaseAsset: BaseAsset{
			SourcePath:  name,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
		template: nil,
	}
	err := m.loadData(isTemplate)
	return m, err
}

func defaultValue(defValue string, val interface{}) string {
	if val == nil {
		return defValue
	}
	strVal, ok := val.(string)
	if !ok || strVal == "" {
		return defValue
	}
	return strVal
}

func (m *BinAsset) loadData(isTemplate bool) error {
	contents, err := Asset(m.SourcePath)
	if err != nil {
		return err
	}

	if isTemplate {
		tpl, err := template.New(m.SourcePath).Funcs(template.FuncMap{"default": defaultValue}).Parse(string(contents))
		if err != nil {
			return err
		}

		m.template = tpl
	}

	m.length = len(contents)
	m.reader = bytes.NewReader(contents)
	glog.V(1).Infof("Created asset %s with %d bytes", m.SourcePath, m.length)
	if m.length == 0 {
		return fmt.Errorf("%s is an empty asset", m.SourcePath)
	}
	return nil
}

// IsTemplate returns if the asset is a template
func (m *BinAsset) IsTemplate() bool {
	return m.template != nil
}

// Evaluate evaluates the template to a new asset
func (m *BinAsset) Evaluate(data interface{}) (*MemoryAsset, error) {
	if !m.IsTemplate() {
		return nil, errors.Errorf("the asset %s is not a template", m.SourcePath)
	}

	var buf bytes.Buffer
	if err := m.template.Execute(&buf, data); err != nil {
		return nil, err
	}

	return NewMemoryAsset(buf.Bytes(), m.GetTargetDir(), m.GetTargetName(), m.GetPermissions()), nil
}

// GetLength returns length
func (m *BinAsset) GetLength() int {
	return m.length
}

// Read reads the asset
func (m *BinAsset) Read(p []byte) (int, error) {
	if m.GetLength() == 0 {
		return 0, fmt.Errorf("attempted read from a 0 length asset")
	}
	return m.reader.Read(p)
}

// Seek resets the reader to offset
func (m *BinAsset) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}
