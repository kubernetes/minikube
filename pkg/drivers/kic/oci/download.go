package oci

import (
	"bytes"
	"os/exec"
)

func PullImage(ociBin, img string) (bytes.Buffer, error) {
	res, err := runCmd(exec.Command(ociBin, "pull", "--quiet", img))
	return res.Stdout, err
}
