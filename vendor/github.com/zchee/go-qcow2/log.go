// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import (
	"bytes"
	"log"
)

func PrintByte(buf []byte) {
	log.Printf("    [0:3] Magic:                       %+v \t %v [81 70 73 251]", buf[:4], bytes.Equal(buf[:4], []byte{81, 70, 73, 251}))
	log.Printf("    [4:7] Version:                     %+v \t\t %v [0 0 0 3]", buf[4:8], bytes.Equal(buf[4:8], []byte{0, 0, 0, 3}))
	log.Printf("   [8:15] BackingFileOffset:           %+v \t %v [0 0 0 0 0 0 0 0]", buf[8:16], bytes.Equal(buf[8:16], []byte{0, 0, 0, 0, 0, 0, 0, 0}))
	log.Printf("  [16:19] BackingFileSize:             %+v \t\t %v [0 0 0 0]", buf[16:20], bytes.Equal(buf[16:20], []byte{0, 0, 0, 0}))
	log.Printf("  [20:23] ClusterBits:                 %+v \t\t %v [0 0 0 16]", buf[20:24], bytes.Equal(buf[20:24], []byte{0, 0, 0, 16}))
	log.Printf("  [24:31] Size:                        %+v \t %v [0 0 0 16 0 0 0 0]", buf[24:32], bytes.Equal(buf[24:32], []byte{0, 0, 0, 16, 0, 0, 0, 0}))
	log.Printf("  [32:35] CryptMethod:                 %+v \t\t %v [0 0 0 0]", buf[32:36], bytes.Equal(buf[32:36], []byte{0, 0, 0, 0}))
	log.Printf("  [36:39] L1Size:                      %+v \t\t %v [0 0 0 128]", buf[36:40], bytes.Equal(buf[36:40], []byte{0, 0, 0, 128}))
	log.Printf("  [40:47] L1TableOffset:               %+v \t %v [0 0 0 0 0 7 0 0]", buf[40:48], bytes.Equal(buf[40:48], []byte{0, 0, 0, 0, 0, 7, 0, 0}))
	log.Printf("  [48:55] RefcountTableOffset:         %+v \t %v [0 0 0 0 0 1 0 0]", buf[48:56], bytes.Equal(buf[48:56], []byte{0, 0, 0, 0, 0, 1, 0, 0}))
	log.Printf("  [56:59] RefcountTableClusters:       %+v \t\t %v [0 0 0 1]", buf[56:60], bytes.Equal(buf[56:60], []byte{0, 0, 0, 1}))
	log.Printf("  [60:63] NbSnapshots:                 %+v \t\t %v [0 0 0 0]", buf[60:64], bytes.Equal(buf[60:64], []byte{0, 0, 0, 0}))
	log.Printf("  [64:71] SnapshotsOffset:             %+v \t %v [0 0 0 0 0 0 0 0]", buf[64:72], bytes.Equal(buf[64:72], []byte{0, 0, 0, 0, 0, 0, 0, 0}))
	log.Printf("  [72:79] IncompatibleFeatures:        %+v \t %v [0 0 0 0 0 0 0 0]", buf[72:80], bytes.Equal(buf[72:80], []byte{0, 0, 0, 0, 0, 0, 0, 0}))
	log.Printf("  [80:87] CompatibleFeatures:          %+v \t %v [0 0 0 0 0 0 0 1]", buf[80:88], bytes.Equal(buf[80:88], []byte{0, 0, 0, 0, 0, 0, 0, 1}))
	log.Printf("  [88:95] AutoclearFeatures:           %+v \t %v [0 0 0 0 0 0 0 0]", buf[88:96], bytes.Equal(buf[88:96], []byte{0, 0, 0, 0, 0, 0, 0, 0}))
	log.Printf("  [96:99] RefcountOrder:               %+v \t\t %v [0 0 0 4]", buf[96:100], bytes.Equal(buf[96:100], []byte{0, 0, 0, 4}))
	log.Printf("[101:105] HeaderLength:                %+v \t\t %v [0 0 0 104]", buf[100:104], bytes.Equal(buf[100:104], []byte{0, 0, 0, 104}))

	log.Printf("[106:109] HeaderExtensionType(type):   %+v \t %v [104 3 248 87]", buf[104:108], bytes.Equal(buf[104:108], []byte{104, 3, 248, 87}))
	log.Printf("[110:114] HeaderExtensionLength(bit):  %+v \t\t %v [0 0 0 144]", buf[108:112], bytes.Equal(buf[108:112], []byte{0, 0, 0, 144}))
	log.Printf("[115:162] HeaderExtensionData(name):   \n%+v\n[0 0 100 105 114 116 121 32 98 105 116 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]", buf[112:158])

	log.Printf("[163:166] HeaderExtensionType(type):   %+v \t\t %v [0 0 0 1]", buf[158:162])
	log.Printf("[167:171] HeaderExtensionLength(bit):  %+v \t %v [99 111 114 114]", buf[162:166])
	log.Printf("[172:219] HeaderExtensionData(name):   \n%+v\n[117 112 116 32 98 105 116 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 108 97]", buf[166:212])

	log.Printf("[220:223] HeaderExtensionType(type):   %+v \t %v [122 121 32 114]", buf[212:216])
	log.Printf("[224:228] HeaderExtensionLength(bit):  %+v \t %v [101 102 99 111]", buf[216:220])
	log.Printf("[229:276] HeaderExtensionData(name):   \n%+v\n[117 110 116 115 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]", buf[220:266])

	if len(buf) > 266 {
		log.Printf("[277:]    Other:                 %+v", buf[277:len(buf)-1])
	}

	log.Printf("byte length: %+v", len(buf))
}
