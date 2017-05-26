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
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

type CopyableFile interface {
	io.Reader
	GetLength() int
	GetAssetName() string
	GetTargetDir() string
	GetTargetName() string
	GetPermissions() string
}

type BaseAsset struct {
	data        []byte
	reader      io.Reader
	Length      int
	AssetName   string
	TargetDir   string
	TargetName  string
	Permissions string
}

func (b *BaseAsset) GetAssetName() string {
	return b.AssetName
}

func (b *BaseAsset) GetTargetDir() string {
	return b.TargetDir
}

func (b *BaseAsset) GetTargetName() string {
	return b.TargetName
}

func (b *BaseAsset) GetPermissions() string {
	return b.Permissions
}

type FileAsset struct {
	BaseAsset
}

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

type MemoryAsset struct {
	BaseAsset
}

func NewMemoryAsset(assetName, targetDir, targetName, permissions string) *MemoryAsset {
	m := &MemoryAsset{
		BaseAsset{
			AssetName:   assetName,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
	}
	m.loadData()
	return m
}

func (m *MemoryAsset) loadData() error {
	contents, err := Asset(m.AssetName)
	if err != nil {
		return err
	}
	m.data = contents
	m.Length = len(contents)
	m.reader = bytes.NewReader(m.data)
	return nil
}

func (m *MemoryAsset) GetLength() int {
	return m.Length
}

func (m *MemoryAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func CopyFileLocal(f CopyableFile) error {
	os.MkdirAll(f.GetTargetDir(), os.ModePerm)
	targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())
	os.Remove(targetPath)
	target, err := os.Create(targetPath)
	defer target.Close()

	perms, err := strconv.Atoi(f.GetPermissions())
	if err != nil {
		return errors.Wrap(err, "Error converting permissions to integer")
	}
	target.Chmod(os.FileMode(perms))
	if err != nil {
		return errors.Wrap(err, "Error changing file permissions")
	}

	_, err = io.Copy(target, f)
	if err != nil {
		return errors.Wrap(err, "Error copying file to target location")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") != "" {
		username := os.Getenv("SUDO_USER")
		if username == "" {
			return nil
		}
		usr, err := user.Lookup(username)
		if err != nil {
			return errors.Wrap(err, "Error looking up user")
		}
		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return errors.Wrapf(err, "Error parsing uid for user: %s", username)
		}
		gid, err := strconv.Atoi(usr.Gid)
		if err != nil {
			return errors.Wrapf(err, "Error parsing gid for user: %s", username)
		}
		if err := os.Chown(targetPath, uid, gid); err != nil {
			return errors.Wrapf(err, "Error changing ownership for: %s", targetPath)
		}
	}
	return nil
}
