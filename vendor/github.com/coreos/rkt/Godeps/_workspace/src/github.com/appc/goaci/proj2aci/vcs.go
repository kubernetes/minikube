package proj2aci

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func repoDirExists(projPath, repoDir string) bool {
	path := filepath.Join(projPath, repoDir)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// getId gets first line of commands output which should hold some VCS
// specific id of current code checkout.
func getId(dir, cmd string, params []string) (string, error) {
	args := []string{cmd}
	args = append(args, params...)
	buffer := new(bytes.Buffer)
	cmdPath, err := exec.LookPath(cmd)
	if err != nil {
		return "", nil
	}
	process := &exec.Cmd{
		Path: cmdPath,
		Args: args,
		Env: []string{
			"PATH=" + os.Getenv("PATH"),
		},
		Dir:    dir,
		Stdout: buffer,
	}
	if err := process.Run(); err != nil {
		return "", err
	}
	output := string(buffer.Bytes())
	if newline := strings.Index(output, "\n"); newline < 0 {
		return output, nil
	} else {
		return output[:newline], nil
	}
}

func getLabelAndId(label, path, cmd string, params []string) (string, string, error) {
	if info, err := getId(path, cmd, params); err != nil {
		return "", "", err
	} else {
		return label, info, nil
	}
}

type VCSInfo interface {
	IsValid(path string) bool
	GetLabelAndId(path string) (string, string, error)
}

type GitInfo struct{}

func (info GitInfo) IsValid(path string) bool {
	return repoDirExists(path, ".git")
}

func (info GitInfo) GetLabelAndId(path string) (string, string, error) {
	return getLabelAndId("git", path, "git", []string{"rev-parse", "HEAD"})
}

type HgInfo struct{}

func (info HgInfo) IsValid(path string) bool {
	return repoDirExists(path, ".hg")
}

func (info HgInfo) GetLabelAndId(path string) (string, string, error) {
	return getLabelAndId("hg", path, "hg", []string{"id", "-i"})
}

type SvnInfo struct{}

func (info SvnInfo) IsValid(path string) bool {
	return repoDirExists(path, ".svn")
}

func (info SvnInfo) GetLabelAndId(path string) (string, string, error) {
	return getLabelAndId("svn", path, "svnversion", []string{})
}

type BzrInfo struct{}

func (info BzrInfo) IsValid(path string) bool {
	return repoDirExists(path, ".bzr")
}

func (info BzrInfo) GetLabelAndId(path string) (string, string, error) {
	return getLabelAndId("bzr", path, "bzr", []string{"revno"})
}

func GetVCSInfo(projPath string) (string, string, error) {
	vcses := []VCSInfo{
		GitInfo{},
		HgInfo{},
		SvnInfo{},
		BzrInfo{},
	}

	for _, vcs := range vcses {
		if vcs.IsValid(projPath) {
			return vcs.GetLabelAndId(projPath)
		}
	}
	return "", "", fmt.Errorf("Unknown code repository in %q", projPath)
}
