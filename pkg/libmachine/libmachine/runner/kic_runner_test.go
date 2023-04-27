package runner

import (
	"os"
	"testing"
)

func TestKICRunner(t *testing.T) {
	t.Parallel()

	t.Run("TestTempDirectory", func(t *testing.T) {
		t.Parallel()

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get user home directory: %v", err)
		}

		tests := []struct {
			isMinikubeSnap bool
			isDockerSnap   bool
			want           string
		}{
			{false, false, ""},
			{true, true, home},
			{false, true, home},
			{true, false, home},
		}

		for _, tt := range tests {
			got, err := tempDirectory(tt.isMinikubeSnap, tt.isDockerSnap)
			if err != nil {
				t.Fatalf("failed to get temp directory: %v", err)
			}

			if got != tt.want {
				t.Errorf("tempDirectory(%t, %t) = %s; want %s", tt.isMinikubeSnap, tt.isDockerSnap, got, tt.want)
			}
		}
	})
}
