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
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// MemorySource is the source name used for in-memory copies
const MemorySource = "memory"

// ReadableFile is something that can be read
type ReadableFile interface {
	io.Reader
	GetLength() int
	GetSourcePath() string

	GetPermissions() string
	GetModTime() (time.Time, error)
	Seek(int64, int) (int64, error)
	Close() error
}

// CopyableFile is something that can be written
type WriteableFile interface {
	io.Writer
	SetLength(int)
	GetTargetPath() string
}

// CopyableFile is something that can be copied
type CopyableFile interface {
	ReadableFile
	WriteableFile
}

type writeFn func(d []byte) (n int, err error)

// BaseCopyableFile is something that can be copied and written
type BaseCopyableFile struct {
	ReadableFile

	writer     writeFn
	length     int
	targetPath string
}

// Write is for write something into the file
func (r *BaseCopyableFile) Write(d []byte) (n int, err error) {
	return r.writer(d)
}

// SetLength is for setting the length
func (r *BaseCopyableFile) SetLength(length int) {
	r.length = length
}

// GetTargetPath returns target path
func (r *BaseCopyableFile) GetTargetPath() string {
	return r.targetPath
}

// NewBaseCopyableFile creates a new instance of BaseCopyableFile
func NewBaseCopyableFile(source ReadableFile, writer writeFn, targetPath string) *BaseCopyableFile {
	return &BaseCopyableFile{
		ReadableFile: source,
		writer:       writer,
		targetPath:   targetPath,
	}
}

type Asset interface {
	CopyableFile
	IsTemplate() bool
	Evaluate(data interface{}) (Asset, error)
}

// BaseAsset is the base asset class
type BaseAsset struct {
	SourcePath  string
	TargetPath  string
	Permissions string
	Source      string
}

// GetSourcePath returns asset name
func (b *BaseAsset) GetSourcePath() string {
	return b.SourcePath
}

// GetTargetPath returns target path
func (b *BaseAsset) GetTargetPath() string {
	return b.TargetPath
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
	writer io.Writer
	file   *os.File // Optional pointer to close file through FileAsset.Close()
}

// NewMemoryAssetTarget creates a new MemoryAsset, with target
func NewMemoryAssetTarget(d []byte, targetPath, permissions string) *MemoryAsset {
	return NewMemoryAsset(d, targetPath, permissions)
}

// NewFileAsset creates a new FileAsset
func NewFileAsset(src, targetPath, permissions string) (*FileAsset, error) {
	klog.V(4).Infof("NewFileAsset: %s -> %s", src, targetPath)

	info, err := os.Stat(src)
	if err != nil {
		return nil, errors.Wrapf(err, "stat")
	}

	if info.Size() == 0 {
		klog.Warningf("NewFileAsset: %s is an empty file!", src)
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}

	return &FileAsset{
		BaseAsset: BaseAsset{
			SourcePath:  src,
			TargetPath:  targetPath,
			Permissions: permissions,
		},
		reader: io.NewSectionReader(f, 0, info.Size()),
		file:   f,
	}, nil
}

// GetLength returns the file length, or 0 (on error)
func (f *FileAsset) GetLength() (flen int) {
	fi, err := os.Stat(f.SourcePath)
	if err != nil {
		klog.Errorf("stat(%q) failed: %v", f.SourcePath, err)
		return 0
	}
	return int(fi.Size())
}

// SetLength sets the file length
func (f *FileAsset) SetLength(flen int) {
	err := os.Truncate(f.SourcePath, int64(flen))
	if err != nil {
		klog.Errorf("truncate(%q) failed: %v", f.SourcePath, err)
	}
}

// GetModTime returns modification time of the file
func (f *FileAsset) GetModTime() (time.Time, error) {
	fi, err := os.Stat(f.SourcePath)
	if err != nil {
		klog.Errorf("stat(%q) failed: %v", f.SourcePath, err)
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

// Write writes the asset
func (f *FileAsset) Write(p []byte) (int, error) {
	if f.writer == nil {
		f.file.Close()
		perms, err := strconv.ParseUint(f.Permissions, 8, 32)
		if err != nil || perms > 07777 {
			return 0, err
		}
		f.file, err = os.OpenFile(f.SourcePath, os.O_RDWR|os.O_CREATE, os.FileMode(perms))
		if err != nil {
			return 0, err
		}
		f.writer = io.Writer(f.file)
	}
	return f.writer.Write(p)
}

// Seek resets the reader to offset
func (f *FileAsset) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

// Close closes the opend file.
func (f *FileAsset) Close() error {
	if f.file == nil {
		return nil
	}
	return f.file.Close()
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

// SetLength returns length
func (m *MemoryAsset) SetLength(len int) {
	m.length = len
}

// Read reads the asset
func (m *MemoryAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

// Writer writes the asset
func (m *MemoryAsset) Write(p []byte) (int, error) {
	m.length = len(p)
	m.reader = bytes.NewReader(p)
	return len(p), nil
}

// Seek resets the reader to offset
func (m *MemoryAsset) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

// Close implemented for CopyableFile interface. Always return nil.
func (m *MemoryAsset) Close() error {
	return nil
}

// NewMemoryAsset creates a new MemoryAsset
func NewMemoryAsset(d []byte, targetPath, permissions string) *MemoryAsset {
	return &MemoryAsset{
		BaseAsset: BaseAsset{
			TargetPath:  targetPath,
			Permissions: permissions,
			SourcePath:  MemorySource,
		},
		reader: bytes.NewReader(d),
		length: len(d),
	}
}

// IsTemplate returns if the asset is a template
func (m *MemoryAsset) IsTemplate() bool {
	return false
}

// Evaluate evaluates the template to a new asset
func (m *MemoryAsset) Evaluate(data interface{}) (Asset, error) {
	return nil, errors.Errorf("the asset %s is not a template", m.SourcePath)
}

// BinAsset is a bindata (binary data) asset
type BinAsset struct {
	embed.FS
	BaseAsset
	reader   io.ReadSeeker
	template *template.Template
	length   int
}

// NewBinAsset creates a new BinAsset
func NewBinAsset(fs embed.FS, name, targetPath, permissions string, isTemplate bool) (*BinAsset, error) {
	m := &BinAsset{
		FS: fs,
		BaseAsset: BaseAsset{
			SourcePath:  name,
			TargetPath:  targetPath,
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
	contents, err := m.FS.ReadFile(m.SourcePath)
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
	klog.V(1).Infof("Created asset %s with %d bytes", m.SourcePath, m.length)
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
func (m *BinAsset) Evaluate(data interface{}) (Asset, error) {
	if !m.IsTemplate() {
		return nil, errors.Errorf("the asset %s is not a template", m.SourcePath)

	}

	var buf bytes.Buffer
	if err := m.template.Execute(&buf, data); err != nil {
		return nil, err
	}

	var asset = NewMemoryAsset(buf.Bytes(), m.GetTargetPath(), m.GetPermissions())

	return asset, nil
}

// GetLength returns length
func (m *BinAsset) GetLength() int {
	return m.length
}

// SetLength sets length
func (m *BinAsset) SetLength(len int) {
	m.length = len
}

// Read reads the asset
func (m *BinAsset) Read(p []byte) (int, error) {
	if m.GetLength() == 0 {
		return 0, fmt.Errorf("attempted read from a 0 length asset")
	}
	return m.reader.Read(p)
}

// Write writes the asset
func (m *BinAsset) Write(p []byte) (int, error) {
	m.length = len(p)
	m.reader = bytes.NewReader(p)
	return len(p), nil
}

// Seek resets the reader to offset
func (m *BinAsset) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

// Close implemented for CopyableFile interface. Always return nil.
func (m *BinAsset) Close() error {
	return nil
}

// CustomAsset is a custom asset on the filesystem
type CustomAsset struct {
	BaseAsset
	FullPath string
	reader   io.ReadSeeker
	template *template.Template
	length   int
}

// NewCustomAsset creates a new CustomAsset
func NewCustomAsset(fullPath, name, targetPath, permissions string, isTemplate bool) (*CustomAsset, error) {
	m := &CustomAsset{
		FullPath: fullPath,
		BaseAsset: BaseAsset{
			SourcePath:  name,
			TargetPath:  targetPath,
			Permissions: permissions,
		},
		template: nil,
	}
	err := m.loadData(isTemplate)
	return m, err
}

func (m *CustomAsset) loadData(isTemplate bool) error {
	contents, err := os.ReadFile(m.FullPath)
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
	klog.V(1).Infof("Created asset %s with %d bytes", m.SourcePath, m.length)
	if m.length == 0 {
		return fmt.Errorf("%s is an empty asset", m.SourcePath)
	}
	return nil
}

// IsTemplate returns if the asset is a template
func (m *CustomAsset) IsTemplate() bool {
	return m.template != nil
}

// Evaluate evaluates the template to a new asset
func (m *CustomAsset) Evaluate(data interface{}) (Asset, error) {
	if !m.IsTemplate() {
		return nil, errors.Errorf("the asset %s is not a template", m.SourcePath)

	}

	var buf bytes.Buffer
	if err := m.template.Execute(&buf, data); err != nil {
		return nil, err
	}

	var asset = NewMemoryAsset(buf.Bytes(), m.GetTargetPath(), m.GetPermissions())

	return asset, nil
}

// GetLength returns length
func (m *CustomAsset) GetLength() int {
	return m.length
}

// SetLength sets length
func (m *CustomAsset) SetLength(len int) {
	m.length = len
}

// Read reads the asset
func (m *CustomAsset) Read(p []byte) (int, error) {
	if m.GetLength() == 0 {
		return 0, fmt.Errorf("attempted read from a 0 length asset")
	}
	return m.reader.Read(p)
}

// Write writes the asset
func (m *CustomAsset) Write(p []byte) (int, error) {
	m.length = len(p)
	m.reader = bytes.NewReader(p)
	return len(p), nil
}

// Seek resets the reader to offset
func (m *CustomAsset) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

// Close implemented for CopyableFile interface. Always return nil.
func (m *CustomAsset) Close() error {
	return nil
}
