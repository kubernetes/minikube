// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import (
	"syscall"

	"github.com/zchee/go-qcow2/internal/mem"
)

const DEBUG_ALLOC2 = false

func growL1Table(bs *BlockDriverState, minSize uint64, exactSize bool) error {
	s := bs.Opaque
	var newL1Size int64

	if minSize <= uint64(s.L1Size) {
		return nil
	}

	// Do a sanity check on min_size before trying to calculate new_l1_size
	// (this prevents overflows during the while loop for the calculation of
	// new_l1_size)
	if minSize > INT_MAX/UINT64_SIZE {
		return syscall.EFBIG
	}

	if exactSize {
		newL1Size = int64(minSize)
	} else {
		// Bump size up to reduce the number of times we have to grow
		newL1Size = int64(s.L1Size)
		if newL1Size == 0 {
			newL1Size = 1
		}
		for minSize > uint64(newL1Size) {
			newL1Size = (newL1Size*3 + 1) / 2
		}
	}

	if newL1Size > MAX_L1_SIZE/UINT64_SIZE {
		return syscall.EFBIG
	}

	newL1Size2 := UINT64_SIZE * newL1Size
	align := bdrvOptMemAlign(bs)
	newL1Table := posixMemalign(uint64(align), uint64(alignOffset(newL1Size2, 512)))
	if newL1Table == nil {
		return syscall.ENOMEM
	}
	mem.Set(newL1Table, byte(alignOffset(newL1Size2, 512)))

	mem.Cpy(newL1Table, []byte{byte(s.L1Table)}, uintptr(s.L1Size*UINT64_SIZE))

	// write new table (align to cluster)
	// newL1TableOffset, err := AllocClusters(bs, uint64(newL1Size2))
	// if err != nil {
	// 	return err
	// }

	// ret = qcow2_cache_flush(bs, s->refcount_block_cache);
	// if (ret < 0) {
	// 	goto fail;
	// }
	//
	// /* the L1 position has not yet been updated, so these clusters must
	// * indeed be completely free */
	// ret = qcow2_pre_write_overlap_check(bs, 0, new_l1_table_offset,
	// new_l1_size2);
	// if (ret < 0) {
	// 	goto fail;
	// }
	//
	// BLKDBG_EVENT(bs->file, BLKDBG_L1_GROW_WRITE_TABLE);
	// for(i = 0; i < s->l1_size; i++)
	// new_l1_table[i] = cpu_to_be64(new_l1_table[i]);
	// ret = bdrv_pwrite_sync(bs->file, new_l1_table_offset,
	// new_l1_table, new_l1_size2);
	// if (ret < 0)
	// goto fail;
	// for(i = 0; i < s->l1_size; i++)
	// new_l1_table[i] = be64_to_cpu(new_l1_table[i]);
	//
	// /* set new table */
	// BLKDBG_EVENT(bs->file, BLKDBG_L1_GROW_ACTIVATE_TABLE);
	// stl_be_p(data, new_l1_size);
	// stq_be_p(data + 4, new_l1_table_offset);
	// ret = bdrv_pwrite_sync(bs->file, offsetof(QCowHeader, l1_size),
	// data, sizeof(data));
	// if (ret < 0) {
	// 	goto fail;
	// }
	// qemu_vfree(s->l1_table);
	// old_l1_table_offset = s->l1_table_offset;
	// s->l1_table_offset = new_l1_table_offset;
	// s->l1_table = new_l1_table;
	// old_l1_size = s->l1_size;
	// s->l1_size = new_l1_size;
	// qcow2_free_clusters(bs, old_l1_table_offset, old_l1_size * sizeof(uint64_t),
	// QCOW2_DISCARD_OTHER);
	// return 0;
	// fail:
	// qemu_vfree(new_l1_table);
	// qcow2_free_clusters(bs, new_l1_table_offset, new_l1_size2,
	// QCOW2_DISCARD_OTHER);
	// return ret;
	return nil
}
