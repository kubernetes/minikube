// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import (
	"bytes"
	"errors"
	"syscall"
)

func getRefcountRO4(refcountArray *[8][]byte, index uint8) uint64 {
	// TODO(zchee): always uint64(1)
	return uint64(1)
}

// ((uint16_t *)refcount_array)[index] = cpu_to_be16(value);
func setRefcountRO4(refcountArray *[8][]byte, index uint8, value uint64) {
	refcountArray[index] = BEUvarint16(uint16(value))
}

func getRefcount(bs *BlockDriverState, clusterIndex uint64) (uint64, error) {
	s := bs.Opaque

	refcountTableIndex := clusterIndex >> uint(s.RefcountBlockBits)
	if uint32(refcountTableIndex) >= s.RefcountTableSize {
		return 0, nil
	}
	refcountBlockOffset := BEUint64(s.RefcountTable[refcountTableIndex]) & REFT_OFFSET_MASK
	if refcountBlockOffset != 0 {
		return 0, nil
	}

	if offsetIntoCluster(s, int64(refcountBlockOffset)) == 0 {
		// TODO(zchee): implements qcow2_signal_corruption
		// qcow2_signal_corruption(bs, true, -1, -1, "Refblock offset %#" PRIx64
		//                         " unaligned (reftable index: %#" PRIx64 ")",
		//                         refcount_block_offset, refcount_table_index);
		return 0, syscall.EIO
	}

	// TODO(zchee): implements qcow2_cache_get
	// ret = qcow2_cache_get(bs, s->refcount_block_cache, refcount_block_offset, &refcount_block);
	// if (ret < 0) {
	//     return ret;
	// }

	blockIndex := clusterIndex & uint64(s.RefcountBlockSize-1)
	var refcountBlock int
	refcount := s.GetRefcount(refcountBlock, blockIndex)

	// TODO(zchee): implements qcow2_cache_put
	// qcow2_cache_put(bs, s->refcount_block_cache, &refcount_block);

	return refcount, nil
}

func AllocClusters(bs *BlockDriverState, size uint64) (int64, error) {
	var (
		offset int64
		err    error
	)

	for {
		offset, err = AllocClustersNoref(bs, size)
		if err != nil || offset < 0 {
			return offset, err
		}

		err = updateRefcount(bs, offset, int64(size), 1, false, DISCARD_NEVER)
		if err != syscall.EAGAIN {
			break
		}
	}

	return offset, nil
}

func AllocClustersNoref(bs *BlockDriverState, size uint64) (int64, error) {
	s := bs.Opaque

	// We can't allocate clusters if they may still be queued for discard
	if s.CacheDiscards {
		// TODO(zchee): implements
		// void qcow2_process_discards(BlockDriverState *bs, int ret)
	}

	nbClusters := sizeToClusters(s, size)
retry:
	for i := uint64(0); i < nbClusters; i++ {
		s.FreeClusterIndex++
		nextClusterIndex := s.FreeClusterIndex
		refcount, err := getRefcount(bs, nextClusterIndex)

		if err != nil {
			return 0, err
		}
		if refcount != 0 {
			goto retry
		}
	}

	if s.FreeClusterIndex > 0 && s.FreeClusterIndex-1 > (INT64_MAX>>uint(s.ClusterBits)) {
		err := errors.New("File too large")
		return 0, err
	}

	return int64((s.FreeClusterIndex - nbClusters) << uint64(s.ClusterBits)), nil
}

var refcountBlock = [8][]byte{}

func updateRefcount(bs *BlockDriverState, offset, length int64, addend uint64, decrease bool, typ DiscardType) error {
	s := bs.Opaque
	start := startOfCluster(int64(s.ClusterSize), offset)
	last := startOfCluster(int64(s.ClusterSize), offset+length-1)
	// log.Printf("start: %+v last: %+v\n", start, last)

	for clusterOffset := start; clusterOffset <= last; clusterOffset += int64(s.ClusterSize) {
		refcountBlockBits := s.ClusterBits - (s.RefcountOrder - 3)
		refcountBlockSize := 1 << uint(refcountBlockBits)
		clusterIndex := clusterOffset >> uint(s.ClusterBits)

		blockIndex := clusterIndex & int64(refcountBlockSize-1)

		refcount := getRefcountRO4(&refcountBlock, uint8(blockIndex))

		setRefcountRO4(&refcountBlock, uint8(blockIndex), refcount)
	}
	// TODO(zchee): hardcoded 131072
	writeFile(bs, 131072, bytes.Join(refcountBlock[:], nil), int(offset+length))

	return nil
}
