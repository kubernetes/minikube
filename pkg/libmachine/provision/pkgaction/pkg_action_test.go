package pkgaction

import "testing"

func TestActionValue(t *testing.T) {
	if Install.String() != "install" {
		t.Fatalf("Expected %q but got %q", "install", Install.String())
	}
}
