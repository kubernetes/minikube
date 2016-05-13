package provision

import (
	"regexp"
	"testing"
)

func TestOracleLinuxGenerateYumRepoList(t *testing.T) {
	info := &OsRelease{
		ID:      "ol",
		Version: "7.2",
	}
	p := NewOracleLinuxProvisioner(nil)
	p.SetOsReleaseInfo(info)

	buf, err := generateYumRepoList(p)
	if err != nil {
		t.Fatal(err)
	}

	m, err := regexp.MatchString(".*oraclelinux/7.*", buf.String())
	if err != nil {
		t.Fatal(err)
	}

	if !m {
		t.Fatalf("expected match for oraclelinux/7")
	}
}
