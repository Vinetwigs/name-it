// Package history keeps a rolling list of the worst filenames seen during the
// current run, newest first.
package history

import (
	"sync"
	"time"
)

type Crime struct {
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
	Score     int       `json:"score"`
	Severity  string    `json:"severity"`
	Action    string    `json:"action"`
}

type Store struct {
	mu     sync.Mutex
	crimes []Crime
}

func New() *Store { return &Store{} }
func (s *Store) Add(c Crime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.crimes = append([]Crime{c}, s.crimes...)
	if len(s.crimes) > 100 {
		s.crimes = s.crimes[:100]
	}
}
func (s *Store) List(n int) []Crime {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n > len(s.crimes) {
		n = len(s.crimes)
	}
	out := make([]Crime, n)
	copy(out, s.crimes[:n])
	return out
}
