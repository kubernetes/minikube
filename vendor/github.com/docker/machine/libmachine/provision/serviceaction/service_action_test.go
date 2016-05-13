package serviceaction

import "testing"

func TestActionValue(t *testing.T) {
	if Restart.String() != "restart" {
		t.Fatalf("Expected %s but got %s", "install", Restart.String())
	}
}
