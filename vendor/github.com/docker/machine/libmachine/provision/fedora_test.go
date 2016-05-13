package provision

import (
	"regexp"
	"testing"
)

func TestFedoraGenerateYumRepoList(t *testing.T) {
	info := &OsRelease{
		ID: "fedora",
	}
	p := NewFedoraProvisioner(nil)
	p.SetOsReleaseInfo(info)

	buf, err := generateYumRepoList(p)
	if err != nil {
		t.Fatal(err)
	}

	m, err := regexp.MatchString(".*fedora/22.*", buf.String())
	if err != nil {
		t.Fatal(err)
	}

	if !m {
		t.Fatalf("expected match for fedora/22")
	}
}
