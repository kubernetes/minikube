//+build windows

package acl

import (
	"github.com/hectane/go-acl/api"
	"golang.org/x/sys/windows"

	"unsafe"
)

// Create an ExplicitAccess instance granting permissions to the provided SID.
func GrantSid(accessPermissions uint32, sid *windows.SID) api.ExplicitAccess {
	return api.ExplicitAccess{
		AccessPermissions: accessPermissions,
		AccessMode:        api.GRANT_ACCESS,
		Inheritance:       api.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: api.Trustee{
			TrusteeForm: api.TRUSTEE_IS_SID,
			Name:        (*uint16)(unsafe.Pointer(sid)),
		},
	}
}

// Create an ExplicitAccess instance granting permissions to the provided name.
func GrantName(accessPermissions uint32, name string) api.ExplicitAccess {
	return api.ExplicitAccess{
		AccessPermissions: accessPermissions,
		AccessMode:        api.GRANT_ACCESS,
		Inheritance:       api.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: api.Trustee{
			TrusteeForm: api.TRUSTEE_IS_NAME,
			Name:        windows.StringToUTF16Ptr(name),
		},
	}
}

// Create an ExplicitAccess instance denying permissions to the provided SID.
func DenySid(accessPermissions uint32, sid *windows.SID) api.ExplicitAccess {
	return api.ExplicitAccess{
		AccessPermissions: accessPermissions,
		AccessMode:        api.DENY_ACCESS,
		Inheritance:       api.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: api.Trustee{
			TrusteeForm: api.TRUSTEE_IS_SID,
			Name:        (*uint16)(unsafe.Pointer(sid)),
		},
	}
}

// Create an ExplicitAccess instance denying permissions to the provided name.
func DenyName(accessPermissions uint32, name string) api.ExplicitAccess {
	return api.ExplicitAccess{
		AccessPermissions: accessPermissions,
		AccessMode:        api.DENY_ACCESS,
		Inheritance:       api.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: api.Trustee{
			TrusteeForm: api.TRUSTEE_IS_NAME,
			Name:        windows.StringToUTF16Ptr(name),
		},
	}
}
