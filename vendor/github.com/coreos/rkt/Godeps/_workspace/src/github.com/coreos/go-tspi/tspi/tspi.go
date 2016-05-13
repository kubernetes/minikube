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
// #cgo LDFLAGS: -ltspi
import "C"
import "errors"
import "fmt"

func tspiError(tssRet C.TSS_RESULT) error {
	ret := (int)(tssRet)
	if ret == 0 {
		return nil
	}
	if (ret & 0xf000) != 0 {
		ret &= ^(0xf000)
		switch {
		case ret == C.TSS_E_FAIL:
			return errors.New("TSS_E_FAIL")
		case ret == C.TSS_E_BAD_PARAMETER:
			return errors.New("TSS_E_BAD_PARAMETER")
		case ret == C.TSS_E_INTERNAL_ERROR:
			return errors.New("TSS_E_INTERNAL_ERROR")
		case ret == C.TSS_E_OUTOFMEMORY:
			return errors.New("TSS_E_OUTOFMEMORY")
		case ret == C.TSS_E_NOTIMPL:
			return errors.New("TSS_E_NOTIMPL")
		case ret == C.TSS_E_KEY_ALREADY_REGISTERED:
			return errors.New("TSS_E_KEY_ALREADY_REGISTERED")
		case ret == C.TSS_E_TPM_UNEXPECTED:
			return errors.New("TSS_E_TPM_UNEXPECTED")
		case ret == C.TSS_E_COMM_FAILURE:
			return errors.New("TSS_E_COMM_FAILURE")
		case ret == C.TSS_E_TIMEOUT:
			return errors.New("TSS_E_TIMEOUT")
		case ret == C.TSS_E_TPM_UNSUPPORTED_FEATURE:
			return errors.New("TSS_E_TPM_UNSUPPORTED_FEATURE")
		case ret == C.TSS_E_CANCELED:
			return errors.New("TSS_E_CANCELED")
		case ret == C.TSS_E_PS_KEY_NOTFOUND:
			return errors.New("TSS_E_PS_KEY_NOTFOUND")
		case ret == C.TSS_E_PS_KEY_EXISTS:
			return errors.New("TSS_E_PS_KEY_EXISTS")
		case ret == C.TSS_E_PS_BAD_KEY_STATE:
			return errors.New("TSS_E_PS_BAD_KEY_STATE")
		case ret == C.TSS_E_INVALID_OBJECT_TYPE:
			return errors.New("TSS_E_INVALID_OBJECT_TYPE")
		case ret == C.TSS_E_NO_CONNECTION:
			return errors.New("TSS_E_NO_CONNECTION")
		case ret == C.TSS_E_CONNECTION_FAILED:
			return errors.New("TSS_E_CONNECTION_FAILED")
		case ret == C.TSS_E_CONNECTION_BROKEN:
			return errors.New("TSS_E_CONNECTION_BROKEN")
		case ret == C.TSS_E_HASH_INVALID_ALG:
			return errors.New("TSS_E_HASH_INVALID_ALG")
		case ret == C.TSS_E_HASH_INVALID_LENGTH:
			return errors.New("TSS_E_HASH_INVALID_LENGTH")
		case ret == C.TSS_E_HASH_NO_DATA:
			return errors.New("TSS_E_HASH_NO_DATA")
		case ret == C.TSS_E_INVALID_ATTRIB_FLAG:
			return errors.New("TSS_E_INVALID_ATTRIB_FLAG")
		case ret == C.TSS_E_INVALID_ATTRIB_SUBFLAG:
			return errors.New("TSS_E_INVALID_ATTRIB_SUBFLAG")
		case ret == C.TSS_E_INVALID_ATTRIB_DATA:
			return errors.New("TSS_E_INVALID_ATTRIB_DATA")
		case ret == C.TSS_E_INVALID_OBJECT_INITFLAG:
			return errors.New("TSS_E_INVALID_OBJECT_INITFLAG")
		case ret == C.TSS_E_NO_PCRS_SET:
			return errors.New("TSS_E_NO_PCRS_SET")
		case ret == C.TSS_E_KEY_NOT_LOADED:
			return errors.New("TSS_E_KEY_NOT_LOADED")
		case ret == C.TSS_E_KEY_NOT_SET:
			return errors.New("TSS_E_KEY_NOT_SET")
		case ret == C.TSS_E_VALIDATION_FAILED:
			return errors.New("TSS_E_VALIDATION_FAILED")
		case ret == C.TSS_E_TSP_AUTHREQUIRED:
			return errors.New("TSS_E_TSP_AUTHREQUIRED")
		case ret == C.TSS_E_TSP_AUTH2REQUIRED:
			return errors.New("TSS_E_TSP_AUTH2REQUIRED")
		case ret == C.TSS_E_TSP_AUTHFAIL:
			return errors.New("TSS_E_TSP_AUTHFAIL")
		case ret == C.TSS_E_TSP_AUTH2FAIL:
			return errors.New("TSS_E_TSP_AUTH2FAIL")
		case ret == C.TSS_E_KEY_NO_MIGRATION_POLICY:
			return errors.New("TSS_E_KEY_NO_MIGRATION_POLICY")
		case ret == C.TSS_E_POLICY_NO_SECRET:
			return errors.New("TSS_E_POLICY_NO_SECRET")
		case ret == C.TSS_E_INVALID_OBJ_ACCESS:
			return errors.New("TSS_E_INVALID_OBJ_ACCESS")
		case ret == C.TSS_E_INVALID_ENCSCHEME:
			return errors.New("TSS_E_INVALID_ENCSCHEME")
		case ret == C.TSS_E_INVALID_SIGSCHEME:
			return errors.New("TSS_E_INVALID_SIGSCHEME")
		case ret == C.TSS_E_ENC_INVALID_LENGTH:
			return errors.New("TSS_E_ENC_INVALID_LENGTH")
		case ret == C.TSS_E_ENC_NO_DATA:
			return errors.New("TSS_E_ENC_NO_DATA")
		case ret == C.TSS_E_ENC_INVALID_TYPE:
			return errors.New("TSS_E_ENC_INVALID_TYPE")
		case ret == C.TSS_E_INVALID_KEYUSAGE:
			return errors.New("TSS_E_INVALID_KEYUSAGE")
		case ret == C.TSS_E_VERIFICATION_FAILED:
			return errors.New("TSS_E_VERIFICATION_FAILED")
		case ret == C.TSS_E_HASH_NO_IDENTIFIER:
			return errors.New("TSS_E_HASH_NO_IDENTIFIER")
		case ret == C.TSS_E_INVALID_HANDLE:
			return errors.New("TSS_E_INVALID_HANDLE")
		case ret == C.TSS_E_SILENT_CONTEXT:
			return errors.New("TSS_E_SILENT_CONTEXT")
		case ret == C.TSS_E_EK_CHECKSUM:
			return errors.New("TSS_E_EK_CHECKSUM")
		case ret == C.TSS_E_DELEGATION_NOTSET:
			return errors.New("TSS_E_DELEGATION_NOTSET")
		case ret == C.TSS_E_DELFAMILY_NOTFOUND:
			return errors.New("TSS_E_DELFAMILY_NOTFOUND")
		case ret == C.TSS_E_DELFAMILY_ROWEXISTS:
			return errors.New("TSS_E_DELFAMILY_ROWEXISTS")
		case ret == C.TSS_E_VERSION_MISMATCH:
			return errors.New("TSS_E_VERSION_MISMATCH")
		case ret == C.TSS_E_DAA_AR_DECRYPTION_ERROR:
			return errors.New("TSS_E_DAA_AR_DECRYPTION_ERROR")
		case ret == C.TSS_E_DAA_AUTHENTICATION_ERROR:
			return errors.New("TSS_E_DAA_AUTHENTICATION_ERROR")
		case ret == C.TSS_E_DAA_CHALLENGE_RESPONSE_ERROR:
			return errors.New("TSS_E_DAA_CHALLENGE_RESPONSE_ERROR")
		case ret == C.TSS_E_DAA_CREDENTIAL_PROOF_ERROR:
			return errors.New("TSS_E_DAA_CREDENTIAL_PROOF_ERROR")
		case ret == C.TSS_E_DAA_CREDENTIAL_REQUEST_PROOF_ERROR:
			return errors.New("TSS_E_DAA_CREDENTIAL_REQUEST_PROOF_ERROR")
		case ret == C.TSS_E_DAA_ISSUER_KEY_ERROR:
			return errors.New("TSS_E_DAA_ISSUER_KEY_ERROR")
		case ret == C.TSS_E_DAA_PSEUDONYM_ERROR:
			return errors.New("TSS_E_DAA_PSEUDONYM_ERROR")
		case ret == C.TSS_E_INVALID_RESOURCE:
			return errors.New("TSS_E_INVALID_RESOURCE")
		case ret == C.TSS_E_NV_AREA_EXIST:
			return errors.New("TSS_E_NV_AREA_EXIST")
		case ret == C.TSS_E_NV_AREA_NOT_EXIST:
			return errors.New("TSS_E_NV_AREA_NOT_EXIST")
		case ret == C.TSS_E_TSP_TRANS_AUTHFAIL:
			return errors.New("TSS_E_TSP_TRANS_AUTHFAIL")
		case ret == C.TSS_E_TSP_TRANS_AUTHREQUIRED:
			return errors.New("TSS_E_TSP_TRANS_AUTHREQUIRED")
		case ret == C.TSS_E_TSP_TRANS_NOTEXCLUSIVE:
			return errors.New("TSS_E_TSP_TRANS_NOTEXCLUSIVE")
		case ret == C.TSS_E_TSP_TRANS_FAIL:
			return errors.New("TSS_E_TSP_TRANS_FAIL")
		case ret == C.TSS_E_TSP_TRANS_NO_PUBKEY:
			return errors.New("TSS_E_TSP_TRANS_NO_PUBKEY")
		case ret == C.TSS_E_NO_ACTIVE_COUNTER:
			return errors.New("TSS_E_NO_ACTIVE_COUNTER")
		}
		return fmt.Errorf("Unknown TSS error: %x", ret)
	}

	switch {
	case ret == C.TPM_E_NON_FATAL:
		return errors.New("TPM_E_NON_FATAL")
	case ret == C.TPM_E_AUTHFAIL:
		return errors.New("TPM_E_AUTHFAIL")
	case ret == C.TPM_E_BADINDEX:
		return errors.New("TPM_E_BADINDEX")
	case ret == C.TPM_E_BAD_PARAMETER:
		return errors.New("TPM_E_BAD_PARAMETER")
	case ret == C.TPM_E_AUDITFAILURE:
		return errors.New("TPM_E_AUDITFAILURE")
	case ret == C.TPM_E_CLEAR_DISABLED:
		return errors.New("TPM_E_CLEAR_DISABLED")
	case ret == C.TPM_E_DEACTIVATED:
		return errors.New("TPM_E_DEACTIVATED")
	case ret == C.TPM_E_DISABLED:
		return errors.New("TPM_E_DISABLED")
	case ret == C.TPM_E_DISABLED_CMD:
		return errors.New("TPM_E_DISABLED_CMD")
	case ret == C.TPM_E_FAIL:
		return errors.New("TPM_E_FAIL")
	case ret == C.TPM_E_BAD_ORDINAL:
		return errors.New("TPM_E_BAD_ORDINAL")
	case ret == C.TPM_E_INSTALL_DISABLED:
		return errors.New("TPM_E_INSTALL_DISABLED")
	case ret == C.TPM_E_INVALID_KEYHANDLE:
		return errors.New("TPM_E_INVALID_KEYHANDLE")
	case ret == C.TPM_E_KEYNOTFOUND:
		return errors.New("TPM_E_KEYNOTFOUND")
	case ret == C.TPM_E_INAPPROPRIATE_ENC:
		return errors.New("TPM_E_INAPPROPRIATE_ENC")
	case ret == C.TPM_E_MIGRATEFAIL:
		return errors.New("TPM_E_MIGRATEFAIL")
	case ret == C.TPM_E_INVALID_PCR_INFO:
		return errors.New("TPM_E_INVALID_PCR_INFO")
	case ret == C.TPM_E_NOSPACE:
		return errors.New("TPM_E_NOSPACE")
	case ret == C.TPM_E_NOSRK:
		return errors.New("TPM_E_NOSRK")
	case ret == C.TPM_E_NOTSEALED_BLOB:
		return errors.New("TPM_E_NOTSEALED_BLOB")
	case ret == C.TPM_E_OWNER_SET:
		return errors.New("TPM_E_OWNER_SET")
	case ret == C.TPM_E_RESOURCES:
		return errors.New("TPM_E_RESOURCES")
	case ret == C.TPM_E_SHORTRANDOM:
		return errors.New("TPM_E_SHORTRANDOM")
	case ret == C.TPM_E_SIZE:
		return errors.New("TPM_E_SIZE")
	case ret == C.TPM_E_WRONGPCRVAL:
		return errors.New("TPM_E_WRONGPCRVAL")
	case ret == C.TPM_E_BAD_PARAM_SIZE:
		return errors.New("TPM_E_BAD_PARAM_SIZE")
	case ret == C.TPM_E_SHA_THREAD:
		return errors.New("TPM_E_SHA_THREAD")
	case ret == C.TPM_E_SHA_ERROR:
		return errors.New("TPM_E_SHA_ERROR")
	case ret == C.TPM_E_FAILEDSELFTEST:
		return errors.New("TPM_E_FAILEDSELFTEST")
	case ret == C.TPM_E_AUTH2FAIL:
		return errors.New("TPM_E_AUTH2FAIL")
	case ret == C.TPM_E_BADTAG:
		return errors.New("TPM_E_BADTAG")
	case ret == C.TPM_E_IOERROR:
		return errors.New("TPM_E_IOERROR")
	case ret == C.TPM_E_ENCRYPT_ERROR:
		return errors.New("TPM_E_ENCRYPT_ERROR")
	case ret == C.TPM_E_DECRYPT_ERROR:
		return errors.New("TPM_E_DECRYPT_ERROR")
	case ret == C.TPM_E_INVALID_AUTHHANDLE:
		return errors.New("TPM_E_INVALID_AUTHHANDLE")
	case ret == C.TPM_E_NO_ENDORSEMENT:
		return errors.New("TPM_E_NO_ENDORSEMENT")
	case ret == C.TPM_E_INVALID_KEYUSAGE:
		return errors.New("TPM_E_INVALID_KEYUSAGE")
	case ret == C.TPM_E_WRONG_ENTITYTYPE:
		return errors.New("TPM_E_WRONG_ENTITYTYPE")
	case ret == C.TPM_E_INVALID_POSTINIT:
		return errors.New("TPM_E_INVALID_POSTINIT")
	case ret == C.TPM_E_INAPPROPRIATE_SIG:
		return errors.New("TPM_E_INAPPROPRIATE_SIG")
	case ret == C.TPM_E_BAD_KEY_PROPERTY:
		return errors.New("TPM_E_BAD_KEY_PROPERTY")
	case ret == C.TPM_E_BAD_MIGRATION:
		return errors.New("TPM_E_BAD_MIGRATION")
	case ret == C.TPM_E_BAD_SCHEME:
		return errors.New("TPM_E_BAD_SCHEME")
	case ret == C.TPM_E_BAD_DATASIZE:
		return errors.New("TPM_E_BAD_DATASIZE")
	case ret == C.TPM_E_BAD_MODE:
		return errors.New("TPM_E_BAD_MODE")
	case ret == C.TPM_E_BAD_PRESENCE:
		return errors.New("TPM_E_BAD_PRESENCE")
	case ret == C.TPM_E_BAD_VERSION:
		return errors.New("TPM_E_BAD_VERSION")
	case ret == C.TPM_E_NO_WRAP_TRANSPORT:
		return errors.New("TPM_E_NO_WRAP_TRANSPORT")
	case ret == C.TPM_E_AUDITFAIL_UNSUCCESSFUL:
		return errors.New("TPM_E_AUDITFAIL_UNSUCCESSFUL")
	case ret == C.TPM_E_AUDITFAIL_SUCCESSFUL:
		return errors.New("TPM_E_AUDITFAIL_SUCCESSFUL")
	case ret == C.TPM_E_NOTRESETABLE:
		return errors.New("TPM_E_NOTRESETABLE")
	case ret == C.TPM_E_NOTLOCAL:
		return errors.New("TPM_E_NOTLOCAL")
	case ret == C.TPM_E_BAD_TYPE:
		return errors.New("TPM_E_BAD_TYPE")
	case ret == C.TPM_E_INVALID_RESOURCE:
		return errors.New("TPM_E_INVALID_RESOURCE")
	case ret == C.TPM_E_NOTFIPS:
		return errors.New("TPM_E_NOTFIPS")
	case ret == C.TPM_E_INVALID_FAMILY:
		return errors.New("TPM_E_INVALID_FAMILY")
	case ret == C.TPM_E_NO_NV_PERMISSION:
		return errors.New("TPM_E_NO_NV_PERMISSION")
	case ret == C.TPM_E_REQUIRES_SIGN:
		return errors.New("TPM_E_REQUIRES_SIGN")
	case ret == C.TPM_E_KEY_NOTSUPPORTED:
		return errors.New("TPM_E_KEY_NOTSUPPORTED")
	case ret == C.TPM_E_AUTH_CONFLICT:
		return errors.New("TPM_E_AUTH_CONFLICT")
	case ret == C.TPM_E_AREA_LOCKED:
		return errors.New("TPM_E_AREA_LOCKED")
	case ret == C.TPM_E_BAD_LOCALITY:
		return errors.New("TPM_E_BAD_LOCALITY")
	case ret == C.TPM_E_READ_ONLY:
		return errors.New("TPM_E_READ_ONLY")
	case ret == C.TPM_E_PER_NOWRITE:
		return errors.New("TPM_E_PER_NOWRITE")
	case ret == C.TPM_E_FAMILYCOUNT:
		return errors.New("TPM_E_FAMILYCOUNT")
	case ret == C.TPM_E_WRITE_LOCKED:
		return errors.New("TPM_E_WRITE_LOCKED")
	case ret == C.TPM_E_BAD_ATTRIBUTES:
		return errors.New("TPM_E_BAD_ATTRIBUTES")
	case ret == C.TPM_E_INVALID_STRUCTURE:
		return errors.New("TPM_E_INVALID_STRUCTURE")
	case ret == C.TPM_E_KEY_OWNER_CONTROL:
		return errors.New("TPM_E_KEY_OWNER_CONTROL")
	case ret == C.TPM_E_BAD_COUNTER:
		return errors.New("TPM_E_BAD_COUNTER")
	case ret == C.TPM_E_NOT_FULLWRITE:
		return errors.New("TPM_E_NOT_FULLWRITE")
	case ret == C.TPM_E_CONTEXT_GAP:
		return errors.New("TPM_E_CONTEXT_GAP")
	case ret == C.TPM_E_MAXNVWRITES:
		return errors.New("TPM_E_MAXNVWRITES")
	case ret == C.TPM_E_NOOPERATOR:
		return errors.New("TPM_E_NOOPERATOR")
	case ret == C.TPM_E_RESOURCEMISSING:
		return errors.New("TPM_E_RESOURCEMISSING")
	case ret == C.TPM_E_DELEGATE_LOCK:
		return errors.New("TPM_E_DELEGATE_LOCK")
	case ret == C.TPM_E_DELEGATE_FAMILY:
		return errors.New("TPM_E_DELEGATE_FAMILY")
	case ret == C.TPM_E_DELEGATE_ADMIN:
		return errors.New("TPM_E_DELEGATE_ADMIN")
	case ret == C.TPM_E_TRANSPORT_NOTEXCLUSIVE:
		return errors.New("TPM_E_TRANSPORT_NOTEXCLUSIVE")
	case ret == C.TPM_E_OWNER_CONTROL:
		return errors.New("TPM_E_OWNER_CONTROL")
	case ret == C.TPM_E_DAA_RESOURCES:
		return errors.New("TPM_E_DAA_RESOURCES")
	case ret == C.TPM_E_DAA_INPUT_DATA0:
		return errors.New("TPM_E_DAA_INPUT_DATA0")
	case ret == C.TPM_E_DAA_INPUT_DATA1:
		return errors.New("TPM_E_DAA_INPUT_DATA1")
	case ret == C.TPM_E_DAA_ISSUER_SETTINGS:
		return errors.New("TPM_E_DAA_ISSUER_SETTINGS")
	case ret == C.TPM_E_DAA_TPM_SETTINGS:
		return errors.New("TPM_E_DAA_TPM_SETTINGS")
	case ret == C.TPM_E_DAA_STAGE:
		return errors.New("TPM_E_DAA_STAGE")
	case ret == C.TPM_E_DAA_ISSUER_VALIDITY:
		return errors.New("TPM_E_DAA_ISSUER_VALIDITY")
	case ret == C.TPM_E_DAA_WRONG_W:
		return errors.New("TPM_E_DAA_WRONG_W")
	case ret == C.TPM_E_BAD_HANDLE:
		return errors.New("TPM_E_BAD_HANDLE")
	case ret == C.TPM_E_BAD_DELEGATE:
		return errors.New("TPM_E_BAD_DELEGATE")
	case ret == C.TPM_E_BADCONTEXT:
		return errors.New("TPM_E_BADCONTEXT")
	case ret == C.TPM_E_TOOMANYCONTEXTS:
		return errors.New("TPM_E_TOOMANYCONTEXTS")
	case ret == C.TPM_E_MA_TICKET_SIGNATURE:
		return errors.New("TPM_E_MA_TICKET_SIGNATURE")
	case ret == C.TPM_E_MA_DESTINATION:
		return errors.New("TPM_E_MA_DESTINATION")
	case ret == C.TPM_E_MA_SOURCE:
		return errors.New("TPM_E_MA_SOURCE")
	case ret == C.TPM_E_MA_AUTHORITY:
		return errors.New("TPM_E_MA_AUTHORITY")
	case ret == C.TPM_E_PERMANENTEK:
		return errors.New("TPM_E_PERMANENTEK")
	case ret == C.TPM_E_BAD_SIGNATURE:
		return errors.New("TPM_E_BAD_SIGNATURE")
	case ret == C.TPM_E_NOCONTEXTSPACE:
		return errors.New("TPM_E_NOCONTEXTSPACE")
	case ret == C.TPM_E_RETRY:
		return errors.New("TPM_E_RETRY")
	case ret == C.TPM_E_NEEDS_SELFTEST:
		return errors.New("TPM_E_NEEDS_SELFTEST")
	case ret == C.TPM_E_DOING_SELFTEST:
		return errors.New("TPM_E_DOING_SELFTEST")
	case ret == C.TPM_E_DEFEND_LOCK_RUNNING:
		return errors.New("TPM_E_DEFEND_LOCK_RUNNING")
	}
	return fmt.Errorf("Unknown error: %x", ret)
}

const (
	TSS_OBJECT_TYPE_POLICY                                = 0x01
	TSS_OBJECT_TYPE_RSAKEY                                = 0x02
	TSS_OBJECT_TYPE_ENCDATA                               = 0x03
	TSS_OBJECT_TYPE_PCRS                                  = 0x04
	TSS_OBJECT_TYPE_HASH                                  = 0x05
	TSS_OBJECT_TYPE_DELFAMILY                             = 0x06
	TSS_OBJECT_TYPE_NV                                    = 0x07
	TSS_OBJECT_TYPE_MIGDATA                               = 0x08
	TSS_OBJECT_TYPE_DAA_CERTIFICATE                       = 0x09
	TSS_OBJECT_TYPE_DAA_ISSUER_KEY                        = 0x0a
	TSS_OBJECT_TYPE_DAA_ARA_KEY                           = 0x0b
	TSS_KEY_NO_AUTHORIZATION                              = 0x00000000
	TSS_KEY_AUTHORIZATION                                 = 0x00000001
	TSS_KEY_AUTHORIZATION_PRIV_USE_ONLY                   = 0x00000002
	TSS_KEY_NON_VOLATILE                                  = 0x00000000
	TSS_KEY_VOLATILE                                      = 0x00000004
	TSS_KEY_NOT_MIGRATABLE                                = 0x00000000
	TSS_KEY_MIGRATABLE                                    = 0x00000008
	TSS_KEY_TYPE_DEFAULT                                  = 0x00000000
	TSS_KEY_TYPE_SIGNING                                  = 0x00000010
	TSS_KEY_TYPE_STORAGE                                  = 0x00000020
	TSS_KEY_TYPE_IDENTITY                                 = 0x00000030
	TSS_KEY_TYPE_AUTHCHANGE                               = 0x00000040
	TSS_KEY_TYPE_BIND                                     = 0x00000050
	TSS_KEY_TYPE_LEGACY                                   = 0x00000060
	TSS_KEY_TYPE_MIGRATE                                  = 0x00000070
	TSS_KEY_TYPE_BITMASK                                  = 0x000000F0
	TSS_KEY_SIZE_DEFAULT                                  = 0x00000000
	TSS_KEY_SIZE_512                                      = 0x00000100
	TSS_KEY_SIZE_1024                                     = 0x00000200
	TSS_KEY_SIZE_2048                                     = 0x00000300
	TSS_KEY_SIZE_4096                                     = 0x00000400
	TSS_KEY_SIZE_8192                                     = 0x00000500
	TSS_KEY_SIZE_16384                                    = 0x00000600
	TSS_KEY_SIZE_BITMASK                                  = 0x00000F00
	TSS_KEY_NOT_CERTIFIED_MIGRATABLE                      = 0x00000000
	TSS_KEY_CERTIFIED_MIGRATABLE                          = 0x00001000
	TSS_KEY_STRUCT_DEFAULT                                = 0x00000000
	TSS_KEY_STRUCT_KEY                                    = 0x00004000
	TSS_KEY_STRUCT_KEY12                                  = 0x00008000
	TSS_KEY_STRUCT_BITMASK                                = 0x0001C000
	TSS_KEY_EMPTY_KEY                                     = 0x00000000
	TSS_KEY_TSP_SRK                                       = 0x04000000
	TSS_KEY_TEMPLATE_BITMASK                              = 0xFC000000
	TSS_ENCDATA_SEAL                                      = 0x00000001
	TSS_ENCDATA_BIND                                      = 0x00000002
	TSS_ENCDATA_LEGACY                                    = 0x00000003
	TSS_HASH_DEFAULT                                      = 0x00000000
	TSS_HASH_SHA1                                         = 0x00000001
	TSS_HASH_OTHER                                        = 0xFFFFFFFF
	TSS_POLICY_USAGE                                      = 0x00000001
	TSS_POLICY_MIGRATION                                  = 0x00000002
	TSS_POLICY_OPERATOR                                   = 0x00000003
	TSS_PCRS_STRUCT_DEFAULT                               = 0x00000000
	TSS_PCRS_STRUCT_INFO                                  = 0x00000001
	TSS_PCRS_STRUCT_INFO_LONG                             = 0x00000002
	TSS_PCRS_STRUCT_INFO_SHORT                            = 0x00000003
	TSS_TSPATTRIB_CONTEXT_SILENT_MODE                     = 0x00000001
	TSS_TSPATTRIB_CONTEXT_MACHINE_NAME                    = 0x00000002
	TSS_TSPATTRIB_CONTEXT_VERSION_MODE                    = 0x00000003
	TSS_TSPATTRIB_CONTEXT_TRANSPORT                       = 0x00000004
	TSS_TSPATTRIB_CONTEXT_CONNECTION_VERSION              = 0x00000005
	TSS_TSPATTRIB_SECRET_HASH_MODE                        = 0x00000006
	TSS_TSPATTRIB_CONTEXTTRANS_CONTROL                    = 0x00000008
	TSS_TSPATTRIB_CONTEXTTRANS_MODE                       = 0x00000010
	TSS_TSPATTRIB_CONTEXT_NOT_SILENT                      = 0x00000000
	TSS_TSPATTRIB_CONTEXT_SILENT                          = 0x00000001
	TSS_TSPATTRIB_CONTEXT_VERSION_AUTO                    = 0x00000001
	TSS_TSPATTRIB_CONTEXT_VERSION_V1_1                    = 0x00000002
	TSS_TSPATTRIB_CONTEXT_VERSION_V1_2                    = 0x00000003
	TSS_TSPATTRIB_DISABLE_TRANSPORT                       = 0x00000016
	TSS_TSPATTRIB_ENABLE_TRANSPORT                        = 0x00000032
	TSS_TSPATTRIB_TRANSPORT_NO_DEFAULT_ENCRYPTION         = 0x00000000
	TSS_TSPATTRIB_TRANSPORT_DEFAULT_ENCRYPTION            = 0x00000001
	TSS_TSPATTRIB_TRANSPORT_AUTHENTIC_CHANNEL             = 0x00000002
	TSS_TSPATTRIB_TRANSPORT_EXCLUSIVE                     = 0x00000004
	TSS_TSPATTRIB_TRANSPORT_STATIC_AUTH                   = 0x00000008
	TSS_CONNECTION_VERSION_1_1                            = 0x00000001
	TSS_CONNECTION_VERSION_1_2                            = 0x00000002
	TSS_TSPATTRIB_SECRET_HASH_MODE_POPUP                  = 0x00000001
	TSS_TSPATTRIB_HASH_MODE_NOT_NULL                      = 0x00000000
	TSS_TSPATTRIB_HASH_MODE_NULL                          = 0x00000001
	TSS_TSPATTRIB_TPM_CALLBACK_COLLATEIDENTITY            = 0x00000001
	TSS_TSPATTRIB_TPM_CALLBACK_ACTIVATEIDENTITY           = 0x00000002
	TSS_TSPATTRIB_TPM_ORDINAL_AUDIT_STATUS                = 0x00000003
	TSS_TSPATTRIB_TPM_CREDENTIAL                          = 0x00001000
	TPM_CAP_PROP_TPM_CLEAR_ORDINAL_AUDIT                  = 0x00000000
	TPM_CAP_PROP_TPM_SET_ORDINAL_AUDIT                    = 0x00000001
	TSS_TPMATTRIB_EKCERT                                  = 0x00000001
	TSS_TPMATTRIB_TPM_CC                                  = 0x00000002
	TSS_TPMATTRIB_PLATFORMCERT                            = 0x00000003
	TSS_TPMATTRIB_PLATFORM_CC                             = 0x00000004
	TSS_TSPATTRIB_POLICY_CALLBACK_HMAC                    = 0x00000080
	TSS_TSPATTRIB_POLICY_CALLBACK_XOR_ENC                 = 0x00000100
	TSS_TSPATTRIB_POLICY_CALLBACK_TAKEOWNERSHIP           = 0x00000180
	TSS_TSPATTRIB_POLICY_CALLBACK_CHANGEAUTHASYM          = 0x00000200
	TSS_TSPATTRIB_POLICY_SECRET_LIFETIME                  = 0x00000280
	TSS_TSPATTRIB_POLICY_POPUPSTRING                      = 0x00000300
	TSS_TSPATTRIB_POLICY_CALLBACK_SEALX_MASK              = 0x00000380
	TSS_TSPATTRIB_POLICY_DELEGATION_INFO                  = 0x00000001
	TSS_TSPATTRIB_POLICY_DELEGATION_PCR                   = 0x00000002
	TSS_SECRET_LIFETIME_ALWAYS                            = 0x00000001
	TSS_SECRET_LIFETIME_COUNTER                           = 0x00000002
	TSS_SECRET_LIFETIME_TIMER                             = 0x00000003
	TSS_TSPATTRIB_POLSECRET_LIFETIME_ALWAYS               = TSS_SECRET_LIFETIME_ALWAYS
	TSS_TSPATTRIB_POLSECRET_LIFETIME_COUNTER              = TSS_SECRET_LIFETIME_COUNTER
	TSS_TSPATTRIB_POLSECRET_LIFETIME_TIMER                = TSS_SECRET_LIFETIME_TIMER
	TSS_TSPATTRIB_POLICYSECRET_LIFETIME_ALWAYS            = TSS_SECRET_LIFETIME_ALWAYS
	TSS_TSPATTRIB_POLICYSECRET_LIFETIME_COUNTER           = TSS_SECRET_LIFETIME_COUNTER
	TSS_TSPATTRIB_POLICYSECRET_LIFETIME_TIMER             = TSS_SECRET_LIFETIME_TIMER
	TSS_TSPATTRIB_POLDEL_TYPE                             = 0x00000001
	TSS_TSPATTRIB_POLDEL_INDEX                            = 0x00000002
	TSS_TSPATTRIB_POLDEL_PER1                             = 0x00000003
	TSS_TSPATTRIB_POLDEL_PER2                             = 0x00000004
	TSS_TSPATTRIB_POLDEL_LABEL                            = 0x00000005
	TSS_TSPATTRIB_POLDEL_FAMILYID                         = 0x00000006
	TSS_TSPATTRIB_POLDEL_VERCOUNT                         = 0x00000007
	TSS_TSPATTRIB_POLDEL_OWNERBLOB                        = 0x00000008
	TSS_TSPATTRIB_POLDEL_KEYBLOB                          = 0x00000009
	TSS_TSPATTRIB_POLDELPCR_LOCALITY                      = 0x00000001
	TSS_TSPATTRIB_POLDELPCR_DIGESTATRELEASE               = 0x00000002
	TSS_TSPATTRIB_POLDELPCR_SELECTION                     = 0x00000003
	TSS_DELEGATIONTYPE_NONE                               = 0x00000001
	TSS_DELEGATIONTYPE_OWNER                              = 0x00000002
	TSS_DELEGATIONTYPE_KEY                                = 0x00000003
	TSS_SECRET_MODE_NONE                                  = 0x00000800
	TSS_SECRET_MODE_SHA1                                  = 0x00001000
	TSS_SECRET_MODE_PLAIN                                 = 0x00001800
	TSS_SECRET_MODE_POPUP                                 = 0x00002000
	TSS_SECRET_MODE_CALLBACK                              = 0x00002800
	TSS_TSPATTRIB_ENCDATA_BLOB                            = 0x00000008
	TSS_TSPATTRIB_ENCDATA_PCR                             = 0x00000010
	TSS_TSPATTRIB_ENCDATA_PCR_LONG                        = 0x00000018
	TSS_TSPATTRIB_ENCDATA_SEAL                            = 0x00000020
	TSS_TSPATTRIB_ENCDATABLOB_BLOB                        = 0x00000001
	TSS_TSPATTRIB_ENCDATAPCR_DIGEST_ATCREATION            = 0x00000002
	TSS_TSPATTRIB_ENCDATAPCR_DIGEST_ATRELEASE             = 0x00000003
	TSS_TSPATTRIB_ENCDATAPCR_SELECTION                    = 0x00000004
	TSS_TSPATTRIB_ENCDATAPCR_DIGEST_RELEASE               = TSS_TSPATTRIB_ENCDATAPCR_DIGEST_ATRELEASE
	TSS_TSPATTRIB_ENCDATAPCRLONG_LOCALITY_ATCREATION      = 0x00000005
	TSS_TSPATTRIB_ENCDATAPCRLONG_LOCALITY_ATRELEASE       = 0x00000006
	TSS_TSPATTRIB_ENCDATAPCRLONG_CREATION_SELECTION       = 0x00000007
	TSS_TSPATTRIB_ENCDATAPCRLONG_RELEASE_SELECTION        = 0x00000008
	TSS_TSPATTRIB_ENCDATAPCRLONG_DIGEST_ATCREATION        = 0x00000009
	TSS_TSPATTRIB_ENCDATAPCRLONG_DIGEST_ATRELEASE         = 0x0000000A
	TSS_TSPATTRIB_ENCDATASEAL_PROTECT_MODE                = 0x00000001
	TSS_TSPATTRIB_ENCDATASEAL_NOPROTECT                   = 0x00000000
	TSS_TSPATTRIB_ENCDATASEAL_PROTECT                     = 0x00000001
	TSS_TSPATTRIB_ENCDATASEAL_NO_PROTECT                  = TSS_TSPATTRIB_ENCDATASEAL_NOPROTECT
	TSS_TSPATTRIB_NV_INDEX                                = 0x00000001
	TSS_TSPATTRIB_NV_PERMISSIONS                          = 0x00000002
	TSS_TSPATTRIB_NV_STATE                                = 0x00000003
	TSS_TSPATTRIB_NV_DATASIZE                             = 0x00000004
	TSS_TSPATTRIB_NV_PCR                                  = 0x00000005
	TSS_TSPATTRIB_NVSTATE_READSTCLEAR                     = 0x00100000
	TSS_TSPATTRIB_NVSTATE_WRITESTCLEAR                    = 0x00200000
	TSS_TSPATTRIB_NVSTATE_WRITEDEFINE                     = 0x00300000
	TSS_TSPATTRIB_NVPCR_READPCRSELECTION                  = 0x01000000
	TSS_TSPATTRIB_NVPCR_READDIGESTATRELEASE               = 0x02000000
	TSS_TSPATTRIB_NVPCR_READLOCALITYATRELEASE             = 0x03000000
	TSS_TSPATTRIB_NVPCR_WRITEPCRSELECTION                 = 0x04000000
	TSS_TSPATTRIB_NVPCR_WRITEDIGESTATRELEASE              = 0x05000000
	TSS_TSPATTRIB_NVPCR_WRITELOCALITYATRELEASE            = 0x06000000
	TSS_NV_TPM                                            = 0x80000000
	TSS_NV_PLATFORM                                       = 0x40000000
	TSS_NV_USER                                           = 0x20000000
	TSS_NV_DEFINED                                        = 0x10000000
	TSS_NV_MASK_TPM                                       = 0x80000000
	TSS_NV_MASK_PLATFORM                                  = 0x40000000
	TSS_NV_MASK_USER                                      = 0x20000000
	TSS_NV_MASK_DEFINED                                   = 0x10000000
	TSS_NV_MASK_RESERVED                                  = 0x0f000000
	TSS_NV_MASK_PURVIEW                                   = 0x00ff0000
	TSS_NV_MASK_INDEX                                     = 0x0000ffff
	TSS_NV_INDEX_SESSIONS                                 = 0x00011101
	TSS_MIGATTRIB_MIGRATIONBLOB                           = 0x00000010
	TSS_MIGATTRIB_MIGRATIONTICKET                         = 0x00000020
	TSS_MIGATTRIB_AUTHORITY_DATA                          = 0x00000030
	TSS_MIGATTRIB_MIG_AUTH_DATA                           = 0x00000040
	TSS_MIGATTRIB_TICKET_DATA                             = 0x00000050
	TSS_MIGATTRIB_PAYLOAD_TYPE                            = 0x00000060
	TSS_MIGATTRIB_MIGRATION_XOR_BLOB                      = 0x00000101
	TSS_MIGATTRIB_MIGRATION_REWRAPPED_BLOB                = 0x00000102
	TSS_MIGATTRIB_MIG_MSALIST_PUBKEY_BLOB                 = 0x00000103
	TSS_MIGATTRIB_MIG_AUTHORITY_PUBKEY_BLOB               = 0x00000104
	TSS_MIGATTRIB_MIG_DESTINATION_PUBKEY_BLOB             = 0x00000105
	TSS_MIGATTRIB_MIG_SOURCE_PUBKEY_BLOB                  = 0x00000106
	TSS_MIGATTRIB_MIG_REWRAPPED_BLOB                      = TSS_MIGATTRIB_MIGRATION_REWRAPPED_BLOB
	TSS_MIGATTRIB_MIG_XOR_BLOB                            = TSS_MIGATTRIB_MIGRATION_XOR_BLOB
	TSS_MIGATTRIB_AUTHORITY_DIGEST                        = 0x00000301
	TSS_MIGATTRIB_AUTHORITY_APPROVAL_HMAC                 = 0x00000302
	TSS_MIGATTRIB_AUTHORITY_MSALIST                       = 0x00000303
	TSS_MIGATTRIB_MIG_AUTH_AUTHORITY_DIGEST               = 0x00000401
	TSS_MIGATTRIB_MIG_AUTH_DESTINATION_DIGEST             = 0x00000402
	TSS_MIGATTRIB_MIG_AUTH_SOURCE_DIGEST                  = 0x00000403
	TSS_MIGATTRIB_TICKET_SIG_DIGEST                       = 0x00000501
	TSS_MIGATTRIB_TICKET_SIG_VALUE                        = 0x00000502
	TSS_MIGATTRIB_TICKET_SIG_TICKET                       = 0x00000503
	TSS_MIGATTRIB_TICKET_RESTRICT_TICKET                  = 0x00000504
	TSS_MIGATTRIB_PT_MIGRATE_RESTRICTED                   = 0x00000601
	TSS_MIGATTRIB_PT_MIGRATE_EXTERNAL                     = 0x00000602
	TSS_TSPATTRIB_HASH_IDENTIFIER                         = 0x00001000
	TSS_TSPATTRIB_ALG_IDENTIFIER                          = 0x00002000
	TSS_TSPATTRIB_PCRS_INFO                               = 0x00000001
	TSS_TSPATTRIB_PCRSINFO_PCRSTRUCT                      = 0x00000001
	TSS_TSPATTRIB_DELFAMILY_STATE                         = 0x00000001
	TSS_TSPATTRIB_DELFAMILY_INFO                          = 0x00000002
	TSS_TSPATTRIB_DELFAMILYSTATE_LOCKED                   = 0x00000001
	TSS_TSPATTRIB_DELFAMILYSTATE_ENABLED                  = 0x00000002
	TSS_TSPATTRIB_DELFAMILYINFO_LABEL                     = 0x00000003
	TSS_TSPATTRIB_DELFAMILYINFO_VERCOUNT                  = 0x00000004
	TSS_TSPATTRIB_DELFAMILYINFO_FAMILYID                  = 0x00000005
	TSS_DELEGATE_INCREMENTVERIFICATIONCOUNT               = 1
	TSS_DELEGATE_CACHEOWNERDELEGATION_OVERWRITEEXISTING   = 1
	TSS_TSPATTRIB_DAACRED_COMMIT                          = 0x00000001
	TSS_TSPATTRIB_DAACRED_ATTRIB_GAMMAS                   = 0x00000002
	TSS_TSPATTRIB_DAACRED_CREDENTIAL_BLOB                 = 0x00000003
	TSS_TSPATTRIB_DAACRED_CALLBACK_SIGN                   = 0x00000004
	TSS_TSPATTRIB_DAACRED_CALLBACK_VERIFYSIGNATURE        = 0x00000005
	TSS_TSPATTRIB_DAACOMMIT_NUMBER                        = 0x00000001
	TSS_TSPATTRIB_DAACOMMIT_SELECTION                     = 0x00000002
	TSS_TSPATTRIB_DAACOMMIT_COMMITMENTS                   = 0x00000003
	TSS_TSPATTRIB_DAAATTRIBGAMMAS_BLOB                    = 0xffffffff
	TSS_TSPATTRIB_DAAISSUERKEY_BLOB                       = 0x00000001
	TSS_TSPATTRIB_DAAISSUERKEY_PUBKEY                     = 0x00000002
	TSS_TSPATTRIB_DAAISSUERKEYBLOB_PUBLIC_KEY             = 0x00000001
	TSS_TSPATTRIB_DAAISSUERKEYBLOB_SECRET_KEY             = 0x00000002
	TSS_TSPATTRIB_DAAISSUERKEYBLOB_KEYBLOB                = 0x00000003
	TSS_TSPATTRIB_DAAISSUERKEYBLOB_PROOF                  = 0x00000004
	TSS_TSPATTRIB_DAAISSUERKEYPUBKEY_NUM_ATTRIBS          = 0x00000001
	TSS_TSPATTRIB_DAAISSUERKEYPUBKEY_NUM_PLATFORM_ATTRIBS = 0x00000002
	TSS_TSPATTRIB_DAAISSUERKEYPUBKEY_NUM_ISSUER_ATTRIBS   = 0x00000003
	TSS_TSPATTRIB_DAAARAKEY_BLOB                          = 0x00000001
	TSS_TSPATTRIB_DAAARAKEYBLOB_PUBLIC_KEY                = 0x00000001
	TSS_TSPATTRIB_DAAARAKEYBLOB_SECRET_KEY                = 0x00000002
	TSS_TSPATTRIB_DAAARAKEYBLOB_KEYBLOB                   = 0x00000003
	TSS_FLAG_DAA_PSEUDONYM_PLAIN                          = 0x00000000
	TSS_FLAG_DAA_PSEUDONYM_ENCRYPTED                      = 0x00000001
	TSS_TSPATTRIB_KEY_BLOB                                = 0x00000040
	TSS_TSPATTRIB_KEY_INFO                                = 0x00000080
	TSS_TSPATTRIB_KEY_UUID                                = 0x000000C0
	TSS_TSPATTRIB_KEY_PCR                                 = 0x00000100
	TSS_TSPATTRIB_RSAKEY_INFO                             = 0x00000140
	TSS_TSPATTRIB_KEY_REGISTER                            = 0x00000180
	TSS_TSPATTRIB_KEY_PCR_LONG                            = 0x000001c0
	TSS_TSPATTRIB_KEY_CONTROLBIT                          = 0x00000200
	TSS_TSPATTRIB_KEY_CMKINFO                             = 0x00000400
	TSS_TSPATTRIB_KEYBLOB_BLOB                            = 0x00000008
	TSS_TSPATTRIB_KEYBLOB_PUBLIC_KEY                      = 0x00000010
	TSS_TSPATTRIB_KEYBLOB_PRIVATE_KEY                     = 0x00000028
	TSS_TSPATTRIB_KEYINFO_SIZE                            = 0x00000080
	TSS_TSPATTRIB_KEYINFO_USAGE                           = 0x00000100
	TSS_TSPATTRIB_KEYINFO_KEYFLAGS                        = 0x00000180
	TSS_TSPATTRIB_KEYINFO_AUTHUSAGE                       = 0x00000200
	TSS_TSPATTRIB_KEYINFO_ALGORITHM                       = 0x00000280
	TSS_TSPATTRIB_KEYINFO_SIGSCHEME                       = 0x00000300
	TSS_TSPATTRIB_KEYINFO_ENCSCHEME                       = 0x00000380
	TSS_TSPATTRIB_KEYINFO_MIGRATABLE                      = 0x00000400
	TSS_TSPATTRIB_KEYINFO_REDIRECTED                      = 0x00000480
	TSS_TSPATTRIB_KEYINFO_VOLATILE                        = 0x00000500
	TSS_TSPATTRIB_KEYINFO_AUTHDATAUSAGE                   = 0x00000580
	TSS_TSPATTRIB_KEYINFO_VERSION                         = 0x00000600
	TSS_TSPATTRIB_KEYINFO_CMK                             = 0x00000680
	TSS_TSPATTRIB_KEYINFO_KEYSTRUCT                       = 0x00000700
	TSS_TSPATTRIB_KEYCONTROL_OWNEREVICT                   = 0x00000780
	TSS_TSPATTRIB_KEYINFO_RSA_EXPONENT                    = 0x00001000
	TSS_TSPATTRIB_KEYINFO_RSA_MODULUS                     = 0x00002000
	TSS_TSPATTRIB_KEYINFO_RSA_KEYSIZE                     = 0x00003000
	TSS_TSPATTRIB_KEYINFO_RSA_PRIMES                      = 0x00004000
	TSS_TSPATTRIB_KEYPCR_DIGEST_ATCREATION                = 0x00008000
	TSS_TSPATTRIB_KEYPCR_DIGEST_ATRELEASE                 = 0x00010000
	TSS_TSPATTRIB_KEYPCR_SELECTION                        = 0x00018000
	TSS_TSPATTRIB_KEYREGISTER_USER                        = 0x02000000
	TSS_TSPATTRIB_KEYREGISTER_SYSTEM                      = 0x04000000
	TSS_TSPATTRIB_KEYREGISTER_NO                          = 0x06000000
	TSS_TSPATTRIB_KEYPCRLONG_LOCALITY_ATCREATION          = 0x00040000
	TSS_TSPATTRIB_KEYPCRLONG_LOCALITY_ATRELEASE           = 0x00080000
	TSS_TSPATTRIB_KEYPCRLONG_CREATION_SELECTION           = 0x000C0000
	TSS_TSPATTRIB_KEYPCRLONG_RELEASE_SELECTION            = 0x00100000
	TSS_TSPATTRIB_KEYPCRLONG_DIGEST_ATCREATION            = 0x00140000
	TSS_TSPATTRIB_KEYPCRLONG_DIGEST_ATRELEASE             = 0x00180000
	TSS_TSPATTRIB_KEYINFO_CMK_MA_APPROVAL                 = 0x00000010
	TSS_TSPATTRIB_KEYINFO_CMK_MA_DIGEST                   = 0x00000020
	TSS_KEY_SIZEVAL_512BIT                                = 0x0200
	TSS_KEY_SIZEVAL_1024BIT                               = 0x0400
	TSS_KEY_SIZEVAL_2048BIT                               = 0x0800
	TSS_KEY_SIZEVAL_4096BIT                               = 0x1000
	TSS_KEY_SIZEVAL_8192BIT                               = 0x2000
	TSS_KEY_SIZEVAL_16384BIT                              = 0x4000
	TSS_KEYUSAGE_BIND                                     = 0x00
	TSS_KEYUSAGE_IDENTITY                                 = 0x01
	TSS_KEYUSAGE_LEGACY                                   = 0x02
	TSS_KEYUSAGE_SIGN                                     = 0x03
	TSS_KEYUSAGE_STORAGE                                  = 0x04
	TSS_KEYUSAGE_AUTHCHANGE                               = 0x05
	TSS_KEYUSAGE_MIGRATE                                  = 0x06
	TSS_KEYFLAG_REDIRECTION                               = 0x00000001
	TSS_KEYFLAG_MIGRATABLE                                = 0x00000002
	TSS_KEYFLAG_VOLATILEKEY                               = 0x00000004
	TSS_KEYFLAG_CERTIFIED_MIGRATABLE                      = 0x00000008
	TSS_ALG_RSA                                           = 0x20
	TSS_ALG_DES                                           = 0x21
	TSS_ALG_3DES                                          = 0x22
	TSS_ALG_SHA                                           = 0x23
	TSS_ALG_HMAC                                          = 0x24
	TSS_ALG_AES128                                        = 0x25
	TSS_ALG_AES192                                        = 0x26
	TSS_ALG_AES256                                        = 0x27
	TSS_ALG_XOR                                           = 0x28
	TSS_ALG_MGF1                                          = 0x29
	TSS_ALG_AES                                           = TSS_ALG_AES128
	TSS_ALG_DEFAULT                                       = 0xfe
	TSS_ALG_DEFAULT_SIZE                                  = 0xff
	TSS_SS_NONE                                           = 0x10
	TSS_SS_RSASSAPKCS1V15_SHA1                            = 0x11
	TSS_SS_RSASSAPKCS1V15_DER                             = 0x12
	TSS_SS_RSASSAPKCS1V15_INFO                            = 0x13
	TSS_ES_NONE                                           = 0x10
	TSS_ES_RSAESPKCSV15                                   = 0x11
	TSS_ES_RSAESOAEP_SHA1_MGF1                            = 0x12
	TSS_ES_SYM_CNT                                        = 0x13
	TSS_ES_SYM_OFB                                        = 0x14
	TSS_ES_SYM_CBC_PKCS5PAD                               = 0x15
	TSS_PS_TYPE_USER                                      = 1
	TSS_PS_TYPE_SYSTEM                                    = 2
	TSS_MS_MIGRATE                                        = 0x20
	TSS_MS_REWRAP                                         = 0x21
	TSS_MS_MAINT                                          = 0x22
	TSS_MS_RESTRICT_MIGRATE                               = 0x23
	TSS_MS_RESTRICT_APPROVE_DOUBLE                        = 0x24
	TSS_MS_RESTRICT_MIGRATE_EXTERNAL                      = 0x25
	TSS_KEYAUTH_AUTH_NEVER                                = 0x10
	TSS_KEYAUTH_AUTH_ALWAYS                               = 0x11
	TSS_KEYAUTH_AUTH_PRIV_USE_ONLY                        = 0x12
	TSS_TPMSTATUS_DISABLEOWNERCLEAR                       = 0x00000001
	TSS_TPMSTATUS_DISABLEFORCECLEAR                       = 0x00000002
	TSS_TPMSTATUS_DISABLED                                = 0x00000003
	TSS_TPMSTATUS_DEACTIVATED                             = 0x00000004
	TSS_TPMSTATUS_OWNERSETDISABLE                         = 0x00000005
	TSS_TPMSTATUS_SETOWNERINSTALL                         = 0x00000006
	TSS_TPMSTATUS_DISABLEPUBEKREAD                        = 0x00000007
	TSS_TPMSTATUS_ALLOWMAINTENANCE                        = 0x00000008
	TSS_TPMSTATUS_PHYSPRES_LIFETIMELOCK                   = 0x00000009
	TSS_TPMSTATUS_PHYSPRES_HWENABLE                       = 0x0000000A
	TSS_TPMSTATUS_PHYSPRES_CMDENABLE                      = 0x0000000B
	TSS_TPMSTATUS_PHYSPRES_LOCK                           = 0x0000000C
	TSS_TPMSTATUS_PHYSPRESENCE                            = 0x0000000D
	TSS_TPMSTATUS_PHYSICALDISABLE                         = 0x0000000E
	TSS_TPMSTATUS_CEKP_USED                               = 0x0000000F
	TSS_TPMSTATUS_PHYSICALSETDEACTIVATED                  = 0x00000010
	TSS_TPMSTATUS_SETTEMPDEACTIVATED                      = 0x00000011
	TSS_TPMSTATUS_POSTINITIALISE                          = 0x00000012
	TSS_TPMSTATUS_TPMPOST                                 = 0x00000013
	TSS_TPMSTATUS_TPMPOSTLOCK                             = 0x00000014
	TSS_TPMSTATUS_DISABLEPUBSRKREAD                       = 0x00000016
	TSS_TPMSTATUS_MAINTENANCEUSED                         = 0x00000017
	TSS_TPMSTATUS_OPERATORINSTALLED                       = 0x00000018
	TSS_TPMSTATUS_OPERATOR_INSTALLED                      = TSS_TPMSTATUS_OPERATORINSTALLED
	TSS_TPMSTATUS_FIPS                                    = 0x00000019
	TSS_TPMSTATUS_ENABLEREVOKEEK                          = 0x0000001A
	TSS_TPMSTATUS_ENABLE_REVOKEEK                         = TSS_TPMSTATUS_ENABLEREVOKEEK
	TSS_TPMSTATUS_NV_LOCK                                 = 0x0000001B
	TSS_TPMSTATUS_TPM_ESTABLISHED                         = 0x0000001C
	TSS_TPMSTATUS_RESETLOCK                               = 0x0000001D
	TSS_TPMSTATUS_DISABLE_FULL_DA_LOGIC_INFO              = 0x0000001D
	TSS_TPMCAP_ORD                                        = 0x10
	TSS_TPMCAP_ALG                                        = 0x11
	TSS_TPMCAP_FLAG                                       = 0x12
	TSS_TPMCAP_PROPERTY                                   = 0x13
	TSS_TPMCAP_VERSION                                    = 0x14
	TSS_TPMCAP_VERSION_VAL                                = 0x15
	TSS_TPMCAP_NV_LIST                                    = 0x16
	TSS_TPMCAP_NV_INDEX                                   = 0x17
	TSS_TPMCAP_MFR                                        = 0x18
	TSS_TPMCAP_SYM_MODE                                   = 0x19
	TSS_TPMCAP_HANDLE                                     = 0x1a
	TSS_TPMCAP_TRANS_ES                                   = 0x1b
	TSS_TPMCAP_AUTH_ENCRYPT                               = 0x1c
	TSS_TPMCAP_SET_PERM_FLAGS                             = 0x1d
	TSS_TPMCAP_SET_VENDOR                                 = 0x1e
	TSS_TPMCAP_DA_LOGIC                                   = 0x1f
	TSS_TPMCAP_PROP_PCR                                   = 0x10
	TSS_TPMCAP_PROP_DIR                                   = 0x11
	TSS_TPMCAP_PROP_MANUFACTURER                          = 0x12
	TSS_TPMCAP_PROP_SLOTS                                 = 0x13
	TSS_TPMCAP_PROP_KEYS                                  = TSS_TPMCAP_PROP_SLOTS
	TSS_TPMCAP_PROP_FAMILYROWS                            = 0x14
	TSS_TPMCAP_PROP_DELEGATEROWS                          = 0x15
	TSS_TPMCAP_PROP_OWNER                                 = 0x16
	TSS_TPMCAP_PROP_MAXKEYS                               = 0x18
	TSS_TPMCAP_PROP_AUTHSESSIONS                          = 0x19
	TSS_TPMCAP_PROP_MAXAUTHSESSIONS                       = 0x1a
	TSS_TPMCAP_PROP_TRANSESSIONS                          = 0x1b
	TSS_TPMCAP_PROP_MAXTRANSESSIONS                       = 0x1c
	TSS_TPMCAP_PROP_SESSIONS                              = 0x1d
	TSS_TPMCAP_PROP_MAXSESSIONS                           = 0x1e
	TSS_TPMCAP_PROP_CONTEXTS                              = 0x1f
	TSS_TPMCAP_PROP_MAXCONTEXTS                           = 0x20
	TSS_TPMCAP_PROP_DAASESSIONS                           = 0x21
	TSS_TPMCAP_PROP_MAXDAASESSIONS                        = 0x22
	TSS_TPMCAP_PROP_DAA_INTERRUPT                         = 0x23
	TSS_TPMCAP_PROP_COUNTERS                              = 0x24
	TSS_TPMCAP_PROP_MAXCOUNTERS                           = 0x25
	TSS_TPMCAP_PROP_ACTIVECOUNTER                         = 0x26
	TSS_TPMCAP_PROP_MIN_COUNTER                           = 0x27
	TSS_TPMCAP_PROP_TISTIMEOUTS                           = 0x28
	TSS_TPMCAP_PROP_STARTUPEFFECTS                        = 0x29
	TSS_TPMCAP_PROP_MAXCONTEXTCOUNTDIST                   = 0x2a
	TSS_TPMCAP_PROP_CMKRESTRICTION                        = 0x2b
	TSS_TPMCAP_PROP_DURATION                              = 0x2c
	TSS_TPMCAP_PROP_MAXNVAVAILABLE                        = 0x2d
	TSS_TPMCAP_PROP_INPUTBUFFERSIZE                       = 0x2e
	TSS_TPMCAP_PROP_REVISION                              = 0x2f
	TSS_TPMCAP_PROP_LOCALITIES_AVAIL                      = 0x32
	TSS_RT_KEY                                            = 0x00000010
	TSS_RT_AUTH                                           = 0x00000020
	TSS_RT_TRANS                                          = 0x00000030
	TSS_RT_COUNTER                                        = 0x00000040
	TSS_TCSCAP_ALG                                        = 0x00000001
	TSS_TCSCAP_VERSION                                    = 0x00000002
	TSS_TCSCAP_CACHING                                    = 0x00000003
	TSS_TCSCAP_PERSSTORAGE                                = 0x00000004
	TSS_TCSCAP_MANUFACTURER                               = 0x00000005
	TSS_TCSCAP_PLATFORM_CLASS                             = 0x00000006
	TSS_TCSCAP_TRANSPORT                                  = 0x00000007
	TSS_TCSCAP_PLATFORM_INFO                              = 0x00000008
	TSS_TCSCAP_PROP_KEYCACHE                              = 0x00000100
	TSS_TCSCAP_PROP_AUTHCACHE                             = 0x00000101
	TSS_TCSCAP_PROP_MANUFACTURER_STR                      = 0x00000102
	TSS_TCSCAP_PROP_MANUFACTURER_ID                       = 0x00000103
	TSS_TCSCAP_PLATFORM_VERSION                           = 0x00001100
	TSS_TCSCAP_PLATFORM_TYPE                              = 0x00001101
	TSS_TCSCAP_TRANS_EXCLUSIVE                            = 0x00002100
	TSS_TCSCAP_PROP_HOST_PLATFORM                         = 0x00003001
	TSS_TCSCAP_PROP_ALL_PLATFORMS                         = 0x00003002
	TSS_TSPCAP_ALG                                        = 0x00000010
	TSS_TSPCAP_VERSION                                    = 0x00000011
	TSS_TSPCAP_PERSSTORAGE                                = 0x00000012
	TSS_TSPCAP_MANUFACTURER                               = 0x00000013
	TSS_TSPCAP_RETURNVALUE_INFO                           = 0x00000015
	TSS_TSPCAP_PLATFORM_INFO                              = 0x00000016
	TSS_TSPCAP_PROP_MANUFACTURER_STR                      = 0x00000102
	TSS_TSPCAP_PROP_MANUFACTURER_ID                       = 0x00000103
	TSS_TSPCAP_PLATFORM_TYPE                              = 0x00000201
	TSS_TSPCAP_PLATFORM_VERSION                           = 0x00000202
	TSS_TSPCAP_PROP_RETURNVALUE_INFO                      = 0x00000201
	TSS_EV_CODE_CERT                                      = 0x00000001
	TSS_EV_CODE_NOCERT                                    = 0x00000002
	TSS_EV_XML_CONFIG                                     = 0x00000003
	TSS_EV_NO_ACTION                                      = 0x00000004
	TSS_EV_SEPARATOR                                      = 0x00000005
	TSS_EV_ACTION                                         = 0x00000006
	TSS_EV_PLATFORM_SPECIFIC                              = 0x00000007
	TSS_TSPCAP_RANDOMLIMIT                                = 0x00001000
	TSS_PCRS_DIRECTION_CREATION                           = 1
	TSS_PCRS_DIRECTION_RELEASE                            = 2
	TSS_BLOB_STRUCT_VERSION                               = 0x01
	TSS_BLOB_TYPE_KEY                                     = 0x01
	TSS_BLOB_TYPE_PUBKEY                                  = 0x02
	TSS_BLOB_TYPE_MIGKEY                                  = 0x03
	TSS_BLOB_TYPE_SEALEDDATA                              = 0x04
	TSS_BLOB_TYPE_BOUNDDATA                               = 0x05
	TSS_BLOB_TYPE_MIGTICKET                               = 0x06
	TSS_BLOB_TYPE_PRIVATEKEY                              = 0x07
	TSS_BLOB_TYPE_PRIVATEKEY_MOD1                         = 0x08
	TSS_BLOB_TYPE_RANDOM_XOR                              = 0x09
	TSS_BLOB_TYPE_CERTIFY_INFO                            = 0x0A
	TSS_BLOB_TYPE_KEY_1_2                                 = 0x0B
	TSS_BLOB_TYPE_CERTIFY_INFO_2                          = 0x0C
	TSS_BLOB_TYPE_CMK_MIG_KEY                             = 0x0D
	TSS_BLOB_TYPE_CMK_BYTE_STREAM                         = 0x0E
	TSS_CMK_DELEGATE_SIGNING                              = 1 << 31
	TSS_CMK_DELEGATE_STORAGE                              = 1 << 30
	TSS_CMK_DELEGATE_BIND                                 = 1 << 29
	TSS_CMK_DELEGATE_LEGACY                               = 1 << 28
	TSS_CMK_DELEGATE_MIGRATE                              = 1 << 27
	TSS_DAA_LENGTH_N                                      = 256
	TSS_DAA_LENGTH_F                                      = 13
	TSS_DAA_LENGTH_E                                      = 46
	TSS_DAA_LENGTH_E_PRIME                                = 15
	TSS_DAA_LENGTH_V                                      = 317
	TSS_DAA_LENGTH_SAFETY                                 = 10
	TSS_DAA_LENGTH_HASH                                   = 20
	TSS_DAA_LENGTH_S                                      = 128
	TSS_DAA_LENGTH_GAMMA                                  = 204
	TSS_DAA_LENGTH_RHO                                    = 26
	TSS_DAA_LENGTH_MFG1_GAMMA                             = 214
	TSS_DAA_LENGTH_MGF1_AR                                = 25
	TPM_ALG_RSA                                           = 0x00000001
	TPM_ALG_DES                                           = 0x00000002
	TPM_ALG_3DES                                          = 0x00000003
	TPM_ALG_SHA                                           = 0x00000004
	TPM_ALG_HMAC                                          = 0x00000005
	TPM_ALG_AES                                           = 0x00000006
	TPM_ALG_AES128                                        = TPM_ALG_AES
	TPM_ALG_MGF1                                          = 0x00000007
	TPM_ALG_AES192                                        = 0x00000008
	TPM_ALG_AES256                                        = 0x00000009
	TPM_ALG_XOR                                           = 0x0000000A
	TPM_SS_NONE                                           = 0x0001
	TPM_SS_RSASSAPKCS1v15_SHA1                            = 0x0002
	TPM_SS_RSASSAPKCS1v15_DER                             = 0x0003
	TPM_SS_RSASSAPKCS1v15_INFO                            = 0x0004
	TPM_ES_NONE                                           = 0x0001
	TPM_ES_RSAESPKCSv15                                   = 0x0002
	TPM_ES_RSAESOAEP_SHA1_MGF1                            = 0x0003
	TPM_ES_SYM_CNT                                        = 0x0004
	TPM_ES_SYM_CTR                                        = TPM_ES_SYM_CNT
	TPM_ES_SYM_OFB                                        = 0x0005
	TPM_ES_SYM_CBC_PKCS5PAD                               = 0x00FF
)

var TSS_UUID_SRK = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 1},
}

var TSS_UUID_SK = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 2},
}

var TSS_UUID_RK = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 3},
}

var TSS_UUID_CRK = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 8},
}

var TSS_UUID_USK1 = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 4},
}

var TSS_UUID_USK2 = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 5},
}

var TSS_UUID_USK3 = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 6},
}

var TSS_UUID_USK4 = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 7},
}

var TSS_UUID_USK5 = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 9},
}

var TSS_UUID_USK6 = C.TSS_UUID{
	ulTimeLow:     0,
	usTimeMid:     0,
	usTimeHigh:    0,
	bClockSeqHigh: 0,
	bClockSeqLow:  0,
	rgbNode:       [6]C.BYTE{0, 0, 0, 0, 0, 10},
}
