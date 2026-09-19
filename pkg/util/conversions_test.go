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

package util

import "testing"

func TestConvertMBToBytes(t *testing.T) {
	tests := []struct {
		mbSize int
		want   int64
	}{
		{0, 0},
		{1, 1048576},
		{2500, 2621440000},
		{8192, 8589934592},
	}
	for _, tt := range tests {
		if got := ConvertMBToBytes(tt.mbSize); got != tt.want {
			t.Errorf("ConvertMBToBytes(%d) = %d, want %d", tt.mbSize, got, tt.want)
		}
	}
}

func TestConvertUnsignedBytesToMB(t *testing.T) {
	tests := []struct {
		byteSize uint64
		want     int64
	}{
		{0, 0},
		{1048576, 1},
		{2621440000, 2500},
		{1048575, 0},
		{1572864, 1},
	}
	for _, tt := range tests {
		if got := ConvertUnsignedBytesToMB(tt.byteSize); got != tt.want {
			t.Errorf("ConvertUnsignedBytesToMB(%d) = %d, want %d", tt.byteSize, got, tt.want)
		}
	}
}

func TestConvertBytesToMB(t *testing.T) {
	tests := []struct {
		byteSize int64
		want     int
	}{
		{0, 0},
		{1048576, 1},
		{2621440000, 2500},
		{1048575, 0},
	}
	for _, tt := range tests {
		if got := ConvertBytesToMB(tt.byteSize); got != tt.want {
			t.Errorf("ConvertBytesToMB(%d) = %d, want %d", tt.byteSize, got, tt.want)
		}
	}
}

func TestConvertMBToBytesRoundTrip(t *testing.T) {
	for _, mb := range []int{0, 1, 100, 2500, 8192, 16384} {
		bytes := ConvertMBToBytes(mb)
		if got := ConvertBytesToMB(bytes); got != mb {
			t.Errorf("round trip for %d MB produced %d bytes, converted back to %d MB", mb, bytes, got)
		}
	}
}
