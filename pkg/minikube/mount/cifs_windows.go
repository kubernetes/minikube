// +build windows

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package mount

import (
	"fmt"
	"github.com/danieljoos/wincred"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/minikube/pkg/minikube/drivers/hyperv"
	"k8s.io/minikube/pkg/minikube/out"
	"os"
	"strings"
)

var credentialName = "minikube"

func (m *Cifs) Share() error {
	// Ensure that the current user is administrator because creating a SMB Share requires Administrator privileges.
	err := hyperv.IsWindowsAdministrator()
	if err != nil {
		return errors.Wrap(err,"current user not administrator")
	}

	// Check if Name of the Share already exists or not.
	err = hyperv.PowershellCmd("SmbShare\\Get-SmbShare","-Name",m.HostShareName)
	if err == nil {
		glog.Infof("The share with share name %v already exists. Trying to delete it.", m.HostShareName)
		if err := hyperv.PowershellCmd("SmbShare\\Remove-SmbShare", "-Name", m.HostShareName, "-Force"); err != nil {
			return errors.Wrap(err, "remove smb share")
		}
		glog.Infof("The share with share name %v has been deleted", m.HostShareName)
	}

	// Get the current user so that we can assign full access permissions to only that user.
	user, err := hyperv.CurrentWindowsUser()
	if err != nil {
		return errors.Wrap(err, "get current user")
	}
	glog.Infof("Current User -- [%v]",user)
	glog.Infof("Trying to enable share for CIFS Mounting.")
	if err := hyperv.PowershellCmd("SmbShare\\New-SmbShare", "-Name", m.HostShareName, "-Path", m.HostPath , "-FullAccess", user, "-Temporary"); err != nil {
		return errors.Wrap(err, "new smb share")
	}
	return nil
}

func (m *Cifs) Unshare() error {

	// Ensure that the current user is administrator because creating a SMB Share requires Administrator privileges.
	err := hyperv.IsWindowsAdministrator()
	if err != nil {
		return errors.Wrap(err,"current user not administrator")
	}

	if err := hyperv.PowershellCmd("SmbShare\\Remove-SmbShare", "-Name", m.HostPath, "-Force"); err != nil {
		return errors.Wrap(err,"remove smb share")
	}
	return nil
}

func (m *Cifs) Mount(r mountRunner) error {
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "get hostname")
	}
	user, err := hyperv.CurrentWindowsUser()
	if err != nil {
		return errors.Wrap(err,"get current user")
	}

	// We need the Domain Name of the current user and the User Name. This is used while connecting to the Share from Linux.
	stringSplitResult := strings.Split(user,"\\")
	domain := stringSplitResult[0]
	username := stringSplitResult[1]

	// Only ask the user for the password if it is not found in the Windows Credential Store
	mountPassword,err := getMountPassword()
	if mountPassword == "" || err != nil  {
		err = saveMountPassword()
		if err != nil {
			return errors.Wrap(err, "save password to wincred")
		}
	}

	mountCmd := fmt.Sprintf("sudo mkdir -p %s && sudo mount.cifs //%s/%s %s -o username=%s,password=%s,domain=%s",m.VmDestinationPath, hostname, m.HostShareName, m.VmDestinationPath, username, mountPassword , domain)
	output, err := r.CombinedOutput(mountCmd)
	if err != nil {
		glog.Infof("%s failed: err=%s, output: %q", mountCmd, err, output)
		return errors.Wrap(err, output)
	}
	glog.Infof("%s output: %q", mountCmd, output)
	return nil
}

func (m *Cifs) Unmount(r mountRunner) error {
	// Unmount the minikube destination path
	cmd := umountCmd(m.VmDestinationPath)
	glog.Infof("Will run: %s", cmd)
	output, err := r.CombinedOutput(cmd)
	glog.Infof("unmount force err=%v, out=%s", err, output)
	if err != nil {
		return errors.Wrap(err, output)
	}
	return nil
}

// Save the password for mounting later on.
func saveMountPassword () error {
	user, err := hyperv.CurrentWindowsUser()
	if err != nil {
		return errors.Wrap(err,"get current user")
	}
	out.T(out.Notice,"In order to mount seamlessly from next time onwards, minikube requires your current password.")
	out.T(out.Enabling,"Please type in the password for the user - [{{.username}}]",out.V{"username":user})
	inputPassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))

	if err != nil {
		return errors.Wrap(err, "open terminal")
	}

	cred := wincred.NewGenericCredential(credentialName)
	cred.CredentialBlob = inputPassword
	err = cred.Write()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getMountPassword() (string, error){
	// Check if the Credential exists in the credential store.
	cred, err := wincred.GetGenericCredential(credentialName)
	if err == nil {
		glog.Infof("Credential %v was found in the Windows Credential Store. Using that...",credentialName)
		return string(cred.CredentialBlob), nil
	} else {
		return "", err
	}
}