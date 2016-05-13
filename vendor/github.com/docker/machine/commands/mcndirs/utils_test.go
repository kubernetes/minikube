package mcndirs

import (
	"path"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/mcnutils"
)

func TestGetBaseDir(t *testing.T) {
	// reset any override env var
	BaseDir = ""

	homeDir := mcnutils.GetHomeDir()
	baseDir := GetBaseDir()

	if strings.Index(baseDir, homeDir) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", homeDir, baseDir)
	}
}

func TestGetCustomBaseDir(t *testing.T) {
	root := "/tmp"
	BaseDir = root
	baseDir := GetBaseDir()

	if strings.Index(baseDir, root) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", root, baseDir)
	}
	BaseDir = ""
}

func TestGetMachineDir(t *testing.T) {
	root := "/tmp"
	BaseDir = root
	machineDir := GetMachineDir()

	if strings.Index(machineDir, root) != 0 {
		t.Fatalf("expected machine dir with prefix %s; received %s", root, machineDir)
	}

	path, filename := path.Split(machineDir)
	if strings.Index(path, root) != 0 {
		t.Fatalf("expected base path of %s; received %s", root, path)
	}
	if filename != "machines" {
		t.Fatalf("expected machine dir \"machines\"; received %s", filename)
	}
	BaseDir = ""
}

func TestGetMachineCertDir(t *testing.T) {
	root := "/tmp"
	BaseDir = root
	clientDir := GetMachineCertDir()

	if strings.Index(clientDir, root) != 0 {
		t.Fatalf("expected machine client cert dir with prefix %s; received %s", root, clientDir)
	}

	path, filename := path.Split(clientDir)
	if strings.Index(path, root) != 0 {
		t.Fatalf("expected base path of %s; received %s", root, path)
	}
	if filename != "certs" {
		t.Fatalf("expected machine client dir \"certs\"; received %s", filename)
	}
	BaseDir = ""
}
