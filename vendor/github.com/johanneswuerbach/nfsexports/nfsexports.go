package nfsexports

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	defaultExportsFile = "/etc/exports"
)

// Add export, if exportsFile is an empty string /etc/exports is used
func Add(exportsFile string, identifier string, export string) ([]byte, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	exports, err := ioutil.ReadFile(exportsFile)

	if err != nil {
		if os.IsNotExist(err) {
			exports = []byte{}
		} else {
			return nil, err
		}
	}

	if containsExport(exports, identifier) {
		return exports, nil
	}

	newExports := exports
	if len(newExports) > 0 && !bytes.HasSuffix(exports, []byte("\n")) {
		newExports = append(newExports, '\n')
	}

	newExports = append(newExports, []byte(exportEntry(identifier, export))...)

	if err := verifyNewExports(newExports); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(exportsFile, newExports, 0644); err != nil {
		return nil, err
	}

	return newExports, nil
}

// Remove export, if exportsFile is an empty string /etc/exports is used
func Remove(exportsFile string, identifier string) ([]byte, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	exports, err := ioutil.ReadFile(exportsFile)
	if err != nil {
		return nil, err
	}

	beginMark := []byte(fmt.Sprintf("# BEGIN: %s", identifier))
	endMark := []byte(fmt.Sprintf("# END: %s\n", identifier))

	begin := bytes.Index(exports, beginMark)
	end := bytes.Index(exports, endMark)

	if begin == -1 || end == -1 {
		return nil, fmt.Errorf("Couldn't not find export %s in %s", identifier, exportsFile)
	}

	newExports := append(exports[:begin], exports[end+len(endMark):]...)
	newExports = append(bytes.TrimSpace(newExports), '\n')

	if err := ioutil.WriteFile(exportsFile, newExports, 0644); err != nil {
		return nil, err
	}

	return newExports, nil
}

// ReloadDaemon reload NFS daemon
func ReloadDaemon() error {
	cmd := exec.Command("sudo", "nfsd", "update")
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Reloading nfds failed: %s\n%s", err.Error(), cmd.Stderr)
	}

	return nil
}

func containsExport(exports []byte, identifier string) bool {
	return bytes.Contains(exports, []byte(fmt.Sprintf("# BEGIN: %s\n", identifier)))
}

func exportEntry(identifier string, export string) string {
	return fmt.Sprintf("# BEGIN: %s\n%s\n# END: %s\n", identifier, export, identifier)
}

func verifyNewExports(newExports []byte) error {
	tmpFile, err := ioutil.TempFile("", "exports")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(newExports); err != nil {
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("nfsd", "-F", tmpFile.Name(), "checkexports")
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Export verification failed:\n%s\n%s", cmd.Stderr, err.Error())
	}

	return nil
}
