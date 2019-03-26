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
	"html/template"
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
)

// CopyableFile is something that can be copied
type CopyableFile interface {
	io.Reader
	GetLength() int
	GetAssetName() string
	GetTargetDir() string
	GetTargetName() string
	GetPermissions() string
}

// BaseAsset is the base asset class
type BaseAsset struct {
	data        []byte
	reader      io.Reader
	Length      int
	AssetName   string
	TargetDir   string
	TargetName  string
	Permissions string
}

// GetAssetName returns asset name
func (b *BaseAsset) GetAssetName() string {
	return b.AssetName
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

// FileAsset is an asset using a file
type FileAsset struct {
	BaseAsset
}

// NewMemoryAssetTarget creates a new MemoryAsset, with target
func NewMemoryAssetTarget(d []byte, targetPath, permissions string) *MemoryAsset {
	return NewMemoryAsset(d, path.Dir(targetPath), path.Base(targetPath), permissions)
}

// NewFileAsset creates a new FileAsset
func NewFileAsset(assetName, targetDir, targetName, permissions string) (*FileAsset, error) {
	f := &FileAsset{
		BaseAsset{
			AssetName:   assetName,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
	}
	file, err := os.Open(f.AssetName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error opening file asset: %s", f.AssetName)
	}
	f.reader = file
	return f, nil
}

// GetLength returns the file length, or 0 (on error)
func (f *FileAsset) GetLength() int {
	file, err := os.Open(f.AssetName)
	defer file.Close()
	if err != nil {
		return 0
	}
	fi, err := file.Stat()
	if err != nil {
		return 0
	}
	return int(fi.Size())
}

func (f *FileAsset) Read(p []byte) (int, error) {
	if f.reader == nil {
		return 0, errors.New("Error attempting FileAsset.Read, FileAsset.reader uninitialized")
	}
	return f.reader.Read(p)
}

// MemoryAsset is a memory-based asset
type MemoryAsset struct {
	BaseAsset
}

// GetLength returns length
func (m *MemoryAsset) GetLength() int {
	return m.Length
}

// Read reads the asset
func (m *MemoryAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

// NewMemoryAsset creates a new MemoryAsset
func NewMemoryAsset(d []byte, targetDir, targetName, permissions string) *MemoryAsset {
	m := &MemoryAsset{
		BaseAsset{
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
	}

	m.data = d
	m.Length = len(m.data)
	m.reader = bytes.NewReader(m.data)
	return m
}

// BinDataAsset is a bindata (binary data) asset
type BinDataAsset struct {
	BaseAsset
	template *template.Template
}

// NewBinDataAsset creates a new BinDataAsset
func NewBinDataAsset(assetName, targetDir, targetName, permissions string, isTemplate bool) *BinDataAsset {
	m := &BinDataAsset{
		BaseAsset: BaseAsset{
			AssetName:   assetName,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
		template: nil,
	}
	m.loadData(isTemplate)
	return m
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

func (m *BinDataAsset) loadData(isTemplate bool) error {
	contents, err := Asset(m.AssetName)
	if err != nil {
		return err
	}

	if isTemplate {
		tpl, err := template.New(m.AssetName).Funcs(template.FuncMap{"default": defaultValue}).Parse(string(contents))
		if err != nil {
			return err
		}

		m.template = tpl
	}

	m.data = contents
	m.Length = len(contents)
	m.reader = bytes.NewReader(m.data)
	return nil
}

// IsTemplate returns if the asset is a template
func (m *BinDataAsset) IsTemplate() bool {
	return m.template != nil
}

// Evaluate evaluates the template to a new asset
func (m *BinDataAsset) Evaluate(data interface{}) (*MemoryAsset, error) {
	if !m.IsTemplate() {
		return nil, errors.Errorf("the asset %s is not a template", m.AssetName)

	}

	var buf bytes.Buffer
	if err := m.template.Execute(&buf, data); err != nil {
		return nil, err
	}

	return NewMemoryAsset(buf.Bytes(), m.GetTargetDir(), m.GetTargetName(), m.GetPermissions()), nil
}

// GetLength returns length
func (m *BinDataAsset) GetLength() int {
	return m.Length
}

// Read reads the asset
func (m *BinDataAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}
