package versioncmp

import (
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   int
	}{
		{"1.12", "1.12", 0},
		{"1.0.0", "1", 0},
		{"1", "1.0.0", 0},
		{"1.05.00.0156", "1.0.221.9289", 1},
		{"1", "1.0.1", -1},
		{"1.0.1", "1", 1},
		{"1.0.1", "1.0.2", -1},
		{"1.0.2", "1.0.3", -1},
		{"1.0.3", "1.1", -1},
		{"1.1", "1.1.1", -1},
		{"1.a", "1.b", 0},
		{"1.a", "2.b", -1},
		{"1.1", "1.1.0", 0},
		{"1.1.1", "1.1.2", -1},
		{"1.1.2", "1.2", -1},
		{"1.12.1", "1.13.0-rc1", -1},
		{"1.13.0-rc1", "1.13.0-rc2", -1},
		{"1.13.0-rc1", "1.13.1-rc1", -1},
		{"17.03.0-ce", "17.03.0-ce", 0},
		{"17.03.1-ce", "17.03.2-ce", -1},
		{"17.06.6-ce", "17.09.2-ce", -1},
		{"17.03.0-ce", "17.06.0-ce", -1},
		{"17.03.0-ce-rc2", "17.03.0-ce-rc1", 1},
		{"17.03.0-ce-rc1", "18.03.0-ce-rc1", -1},
		{"17.06.0-ce-rc2", "1.12.0", 1},
		{"1.12.0", "17.06.0-ce-rc2", -1},
	}

	for _, tc := range cases {
		if got := compare(tc.v1, tc.v2); got != tc.want {
			t.Errorf("compare(%q, %q) == %d, want %d", tc.v1, tc.v2, got, tc.want)
		}
	}
}

func TestLessThan(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.12", "1.12", false},
		{"1.0.0", "1", false},
		{"1", "1.0.0", false},
		{"1.05.00.0156", "1.0.221.9289", false},
		{"1", "1.0.1", true},
		{"1.0.1", "1", false},
		{"1.0.1", "1.0.2", true},
		{"1.0.2", "1.0.3", true},
		{"1.0.3", "1.1", true},
		{"1.1", "1.1.1", true},
		{"1.a", "1.b", false},
		{"1.a", "2.b", true},
		{"1.1", "1.1.0", false},
		{"1.1.1", "1.1.2", true},
		{"1.1.2", "1.2", true},
		{"1.12.1", "1.13.0-rc1", true},
		{"1.13.0-rc1", "1.13.0-rc2", true},
		{"1.13.0-rc1", "1.13.1-rc1", true},
		{"17.03.0-ce", "17.03.0-ce", false},
		{"17.03.1-ce", "17.03.2-ce", true},
		{"17.06.6-ce", "17.09.2-ce", true},
		{"17.03.0-ce", "17.06.0-ce", true},
		{"17.03.0-ce-rc2", "17.03.0-ce-rc1", false},
		{"17.03.0-ce-rc1", "18.03.0-ce-rc1", true},
		{"17.06.0-ce", "1.12.0", false},
	}
	for _, tc := range cases {
		if got := LessThan(tc.v1, tc.v2); got != tc.want {
			t.Errorf("LessThan(%q, %q) == %v, want %v", tc.v1, tc.v2, got, tc.want)
		}
	}
}

func TestLessThanOrEqualTo(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.12", "1.12", true},
		{"1.0.0", "1", true},
		{"1", "1.0.0", true},
		{"1.05.00.0156", "1.0.221.9289", false},
		{"1", "1.0.1", true},
		{"1.0.1", "1", false},
		{"1.0.1", "1.0.2", true},
		{"1.0.2", "1.0.3", true},
		{"1.0.3", "1.1", true},
		{"1.1", "1.1.1", true},
		{"1.a", "1.b", true},
		{"1.a", "2.b", true},
		{"1.1", "1.1.0", true},
		{"1.1.1", "1.1.2", true},
		{"1.1.2", "1.2", true},
		{"1.12.1", "1.13.0-rc1", true},
		{"1.13.0-rc1", "1.13.0-rc2", true},
		{"1.13.0-rc1", "1.13.1-rc1", true},
		{"17.03.0-ce", "17.03.0-ce", true},
		{"17.03.1-ce", "17.03.2-ce", true},
		{"17.06.6-ce", "17.09.2-ce", true},
		{"17.03.0-ce", "17.06.0-ce", true},
		{"17.03.0-ce-rc2", "17.03.0-ce-rc1", false},
		{"17.03.0-ce-rc1", "18.03.0-ce-rc1", true},
		{"17.06.0-ce", "1.12.0", false},
	}
	for _, tc := range cases {
		if got := LessThanOrEqualTo(tc.v1, tc.v2); got != tc.want {
			t.Errorf("LessThanOrEqualTo(%q, %q) == %v, want %v", tc.v1, tc.v2, got, tc.want)
		}
	}
}

func TestGreaterThan(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.12", "1.12", false},
		{"1.0.0", "1", false},
		{"1", "1.0.0", false},
		{"1.05.00.0156", "1.0.221.9289", true},
		{"1", "1.0.1", false},
		{"1.0.1", "1", true},
		{"1.0.1", "1.0.2", false},
		{"1.0.2", "1.0.3", false},
		{"1.0.3", "1.1", false},
		{"1.1", "1.1.1", false},
		{"1.a", "1.b", false},
		{"1.a", "2.b", false},
		{"1.1", "1.1.0", false},
		{"1.1.1", "1.1.2", false},
		{"1.1.2", "1.2", false},
		{"1.12.1", "1.13.0-rc1", false},
		{"1.13.0-rc1", "1.13.0-rc2", false},
		{"1.13.0-rc1", "1.13.1-rc1", false},
		{"17.03.0-ce", "17.03.0-ce", false},
		{"17.03.1-ce", "17.03.2-ce", false},
		{"17.06.6-ce", "17.09.2-ce", false},
		{"17.03.0-ce", "17.06.0-ce", false},
		{"17.03.0-ce-rc2", "17.03.0-ce-rc1", true},
		{"17.03.0-ce-rc1", "18.03.0-ce-rc1", false},
		{"17.06.0-ce", "1.12.0", true},
	}
	for _, tc := range cases {
		if got := GreaterThan(tc.v1, tc.v2); got != tc.want {
			t.Errorf("GreaterThan(%q, %q) == %v, want %v", tc.v1, tc.v2, got, tc.want)
		}
	}
}

func TestGreaterThanOrEqualTo(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.12", "1.12", true},
		{"1.0.0", "1", true},
		{"1", "1.0.0", true},
		{"1.05.00.0156", "1.0.221.9289", true},
		{"1", "1.0.1", false},
		{"1.0.1", "1", true},
		{"1.0.1", "1.0.2", false},
		{"1.0.2", "1.0.3", false},
		{"1.0.3", "1.1", false},
		{"1.1", "1.1.1", false},
		{"1.a", "1.b", true},
		{"1.a", "2.b", false},
		{"1.1", "1.1.0", true},
		{"1.1.1", "1.1.2", false},
		{"1.1.2", "1.2", false},
		{"1.12.1", "1.13.0-rc1", false},
		{"1.13.0-rc1", "1.13.0-rc2", false},
		{"1.13.0-rc1", "1.13.1-rc1", false},
		{"17.03.0-ce", "17.03.0-ce", true},
		{"17.03.1-ce", "17.03.2-ce", false},
		{"17.06.6-ce", "17.09.2-ce", false},
		{"17.03.0-ce", "17.06.0-ce", false},
		{"17.03.0-ce-rc2", "17.03.0-ce-rc1", true},
		{"17.03.0-ce-rc1", "18.03.0-ce-rc1", false},
		{"17.06.0-ce", "1.12.0", true},
	}
	for _, tc := range cases {
		if got := GreaterThanOrEqualTo(tc.v1, tc.v2); got != tc.want {
			t.Errorf("GreaterThanOrEqualTo(%q, %q) == %v, want %v", tc.v1, tc.v2, got, tc.want)
		}
	}
}

func TestEqual(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.12", "1.12", true},
		{"1.0.0", "1", true},
		{"1", "1.0.0", true},
		{"1.05.00.0156", "1.0.221.9289", false},
		{"1", "1.0.1", false},
		{"1.0.1", "1", false},
		{"1.0.1", "1.0.2", false},
		{"1.0.2", "1.0.3", false},
		{"1.0.3", "1.1", false},
		{"1.1", "1.1.1", false},
		{"1.a", "1.b", true},
		{"1.a", "2.b", false},
		{"1.1", "1.1.0", true},
		{"1.1.1", "1.1.2", false},
		{"1.1.2", "1.2", false},
		{"1.12.1", "1.13.0-rc1", false},
		{"1.13.0-rc1", "1.13.0-rc2", false},
		{"1.13.0-rc1", "1.13.1-rc1", false},
		{"17.03.0-ce", "17.03.0-ce", true},
		{"17.03.1-ce", "17.03.2-ce", false},
		{"17.06.6-ce", "17.09.2-ce", false},
		{"17.03.0-ce", "17.06.0-ce", false},
		{"17.03.0-ce-rc2", "17.03.0-ce-rc1", false},
		{"17.03.0-ce-rc1", "18.03.0-ce-rc1", false},
		{"17.06.0-ce-rc2", "1.12.0", false},
		{"1.12.0", "17.06.0-ce-rc2", false},
		{"17.06.0-ce", "1.12.0", false},
	}
	for _, tc := range cases {
		if got := Equal(tc.v1, tc.v2); got != tc.want {
			t.Errorf("Equal(%q, %q) == %v, want %v", tc.v1, tc.v2, got, tc.want)
		}
	}
}
