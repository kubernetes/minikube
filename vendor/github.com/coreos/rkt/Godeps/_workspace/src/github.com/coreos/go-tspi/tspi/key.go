// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tspi

// #include <trousers/tss.h>
import "C"
import "unsafe"

// ModulusFromBlob provides the modulus of a provided TSS key blob
func ModulusFromBlob(blob []byte) []byte {
	return blob[28:]
}

// Key is a TSS key
type Key struct {
	handle  C.TSS_HKEY
	context C.TSS_HCONTEXT
}

// GetPolicy returns the policy associated with the key
func (key *Key) GetPolicy(poltype int) (*Policy, error) {
	var policyHandle C.TSS_HPOLICY
	err := tspiError(C.Tspi_GetPolicyObject((C.TSS_HOBJECT)(key.handle), (C.TSS_FLAG)(poltype), &policyHandle))
	return &Policy{handle: policyHandle, context: key.context}, err
}

// SetModulus sets the modulus of a public key to the provided value
func (key *Key) SetModulus(n []byte) error {
	err := tspiError(C.Tspi_SetAttribData((C.TSS_HOBJECT)(key.handle), C.TSS_TSPATTRIB_RSAKEY_INFO, C.TSS_TSPATTRIB_KEYINFO_RSA_MODULUS, (C.UINT32)(len(n)), (*C.BYTE)(unsafe.Pointer(&n[0]))))
	return err
}

// GetModulus returns the modulus of the public key
func (key *Key) GetModulus() (modulus []byte, err error) {
	var dataLen C.UINT32
	var cData *C.BYTE
	err = tspiError(C.Tspi_GetAttribData((C.TSS_HOBJECT)(key.handle), C.TSS_TSPATTRIB_RSAKEY_INFO, C.TSS_TSPATTRIB_KEYINFO_RSA_MODULUS, &dataLen, &cData))
	data := C.GoBytes(unsafe.Pointer(cData), (C.int)(dataLen))
	C.Tspi_Context_FreeMemory(key.context, cData)
	return data, err
}

// GetPubKeyBlob returns the public half of the key in TPM blob format
func (key *Key) GetPubKeyBlob() (pubkey []byte, err error) {
	var dataLen C.UINT32
	var cData *C.BYTE
	err = tspiError(C.Tspi_GetAttribData((C.TSS_HOBJECT)(key.handle), C.TSS_TSPATTRIB_KEY_BLOB, C.TSS_TSPATTRIB_KEYBLOB_PUBLIC_KEY, &dataLen, &cData))
	data := C.GoBytes(unsafe.Pointer(cData), (C.int)(dataLen))
	C.Tspi_Context_FreeMemory(key.context, cData)
	return data, err
}

// GetKeyBlob returns an encrypted blob containing the public and private
// halves of the key
func (key *Key) GetKeyBlob() ([]byte, error) {
	var dataLen C.UINT32
	var cData *C.BYTE
	err := tspiError(C.Tspi_GetAttribData((C.TSS_HOBJECT)(key.handle), C.TSS_TSPATTRIB_KEY_BLOB, C.TSS_TSPATTRIB_KEYBLOB_BLOB, &dataLen, &cData))
	data := C.GoBytes(unsafe.Pointer(cData), (C.int)(dataLen))
	C.Tspi_Context_FreeMemory(key.context, cData)
	return data, err
}

