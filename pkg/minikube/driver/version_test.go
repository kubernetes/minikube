package driver

import (
	"testing"

	"github.com/blang/semver"
)

func Test_minDriverVersion(t *testing.T) {

	tests := []struct {
		desc   string
		driver string
		mkV    string
		want   semver.Version
	}{
		{"Hyperkit", HyperKit, "1.1.1", minHyperkitVersion},
		{"Invalid", "_invalid_", "1.1.1", v("1.1.1")},
		{"KVM2", KVM2, "1.1.1", v("1.1.1")},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := minDriverVersion(tt.driver, v(tt.mkV)); !got.EQ(tt.want) {
				t.Errorf("minDriverVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func v(s string) semver.Version {
	r, err := semver.New(s)
	if err != nil {
		panic(err)
	}
	return *r
}
