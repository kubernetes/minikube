// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import (
	"bytes"
	"os"
)

// BlockOption represents a block options.
type BlockOption struct {
	Driver BlockDriver
}

// BlockBackend represents a backend of the QCow2 image format block driver.
type BlockBackend struct {
	File             *os.File
	Header           Header
	allowBeyondEOF   bool
	BlockDriverState *BlockDriverState

	buf bytes.Buffer

	Error error
}

func (blk *BlockBackend) bs() *BlockDriverState {
	return blk.BlockDriverState
}
