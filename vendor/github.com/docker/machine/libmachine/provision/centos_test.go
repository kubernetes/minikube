package provision

import (
	"regexp"
	"testing"
)

func TestCentosGenerateYumRepoList(t *testing.T) {
	info := &OsRelease{
		ID: "centos",
	}
	p := NewCentosProvisioner(nil)
	p.SetOsReleaseInfo(info)

	buf, err := generateYumRepoList(p)
	if err != nil {
		t.Fatal(err)
	}

	m, err := regexp.MatchString(".*centos/7.*", buf.String())
	if err != nil {
		t.Fatal(err)
	}

	if !m {
		t.Fatalf("expected match for centos/7")
	}
}
