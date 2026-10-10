package watcher

import "testing"

func TestEventTypesAreDistinct(t *testing.T) {
	seen := map[EventType]bool{
		Create: true,
		Rename: true,
		Remove: true,
		Write:  true,
	}

	if len(seen) != 4 {
		t.Fatal("event constants must all hold different values")
	}
}

func TestShouldIgnore(t *testing.T) {
	w := &Watcher{}

	for _, name := range []string{".~lock.report.odt#", "document.autosave", "AUTOSAVE.tmp"} {
		if !w.shouldIgnore(name) {
			t.Errorf("expected %q to be ignored", name)
		}
	}

	for _, name := range []string{"report.txt", "notes.md", "final_final_2.psd"} {
		if w.shouldIgnore(name) {
			t.Errorf("did not expect %q to be ignored", name)
		}
	}
}
