package problem

import (
	"fmt"
	"testing"
)

func TestFromError(t *testing.T) {
	var tests = []struct {
		want string
		err  string
	}{
		{"IP_NOT_FOUND", "bootstrapper: Error creating new ssh host from driver: Error getting ssh host name for driver: IP not found"},
		{"VBOX_HOST_ADAPTER", "Error starting host:  Error starting stopped host: Error setting up host only network on machine start: The host-only adapter we just created is not visible. This is a well known VirtualBox bug. You might want to uninstall it and reinstall at least version 5.0.12 that is is supposed to fix this issue"},
		{"", "xyz"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := FromError(fmt.Errorf(tc.err))
			if got == nil {
				if tc.want != "" {
					t.Errorf("FromError(%q)=nil, want %s", tc.err, tc.want)
				}
				return
			}
			if got.ID != tc.want {
				t.Errorf("FromError(%q)=%s, want %s", tc.err, got.ID, tc.want)
			}
		})
	}
}
