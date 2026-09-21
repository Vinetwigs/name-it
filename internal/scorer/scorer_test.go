package scorer

import (
	"testing"

	"name-it.vino/internal/session"
)

func TestNames(t *testing.T) {
	cases := []struct {
		name     string
		min, max int
	}{
		{"poster_v01.psd", 0, 10}, {"scene03_take02.mov", 0, 10}, {"IMG_4821.jpg", 0, 5},
		{"finale.psd", 5, 30}, {"finale2.psd", 15, 50}, {"finale_finale.psd", 35, 90}, {"finale_finale_2.psd", 50, 100},
		{"copy_copy.psd", 35, 100}, {"poster_2026-09-16.psd", 0, 10},
	}
	for _, tc := range cases {
		a := Assess(tc.name, nil)
		if a.Score < tc.min || a.Score > tc.max {
			t.Errorf("%s score=%d want %d..%d", tc.name, a.Score, tc.min, tc.max)
		}
	}
}
func TestEscalation(t *testing.T) {
	r := []session.FileRecord{{Name: "poster.psd"}, {Name: "poster2.psd"}}
	a := Assess("poster3.psd", r)
	if a.Score < 40 {
		t.Fatalf("expected escalation, got %d", a.Score)
	}
}
