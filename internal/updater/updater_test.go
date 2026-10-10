package updater

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"2.1.0", "2.0.0", true},
		{"2.0.1", "2.0.0", true},
		{"2.0.0", "2.0.0", false},
		{"1.9.9", "2.0.0", false},
		{"2.1.0", "2.1", false},
		{"2.2", "2.1.9", true},
		{"3.0.0-beta.1", "2.9.0", true},
		{"not-a-version", "2.0.0", false},
	}

	for _, tc := range cases {
		if got := IsNewer(tc.candidate, tc.current); got != tc.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tc.candidate, tc.current, got, tc.want)
		}
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want [3]int
	}{
		{"2.1.0", [3]int{2, 1, 0}},
		{"v2.1", [3]int{2, 1, 0}},
		{"2", [3]int{2, 0, 0}},
		{"2.1.0+build.5", [3]int{2, 1, 0}},
		{"2.1.0-rc.1", [3]int{2, 1, 0}},
		{"garbage", [3]int{0, 0, 0}},
	}

	for _, tc := range cases {
		if got := parse(tc.in); got != tc.want {
			t.Errorf("parse(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
