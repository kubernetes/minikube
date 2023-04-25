//+build windows

package acl

import (
	"os"

	"golang.org/x/sys/windows"
)

// Change the permissions of the specified file. Only the nine
// least-significant bytes are used, allowing access by the file's owner, the
// file's group, and everyone else to be explicitly controlled.
func Chmod(name string, fileMode os.FileMode) error {
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	creatorOwnerSID, err := windows.StringToSid("S-1-3-0")
	if err != nil {
		return err
	}
	creatorGroupSID, err := windows.StringToSid("S-1-3-1")
	if err != nil {
		return err
	}
	everyoneSID, err := windows.StringToSid("S-1-1-0")
	if err != nil {
		return err
	}

	mode := uint32(fileMode)
	return Apply(
		name,
		true,
		false,
		GrantSid(((mode&0700)<<23)|((mode&0200)<<9), creatorOwnerSID),
		GrantSid(((mode&0070)<<26)|((mode&0020)<<12), creatorGroupSID),
		GrantSid(((mode&0007)<<29)|((mode&0002)<<15), everyoneSID),
	)
}
