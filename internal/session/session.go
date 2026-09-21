package session

import (
	"path/filepath"
	"sync"
	"time"
)

type FileRecord struct {
	Path      string
	Name      string
	Extension string
	Folder    string
	Timestamp time.Time
}
type State struct {
	Active       bool      `json:"active"`
	Root         string    `json:"root"`
	FilesWatched int       `json:"filesWatched"`
	Questionable int       `json:"questionable"`
	PausedUntil  time.Time `json:"pausedUntil"`
}

type Manager struct {
	mu    sync.Mutex
	state State
	files []FileRecord
}

func New() *Manager { return &Manager{} }
func (m *Manager) Start(root string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = State{Active: true, Root: root}
}
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Active = false
	m.state.PausedUntil = time.Time{}
}
func (m *Manager) Pause(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.PausedUntil = time.Now().Add(d)
}
func (m *Manager) Paused() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.state.PausedUntil.IsZero() && time.Now().Before(m.state.PausedUntil)
}
func (m *Manager) Record(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files = append(m.files, FileRecord{Path: path, Name: filepath.Base(path), Extension: filepath.Ext(path), Folder: filepath.Dir(path), Timestamp: time.Now()})
	m.state.FilesWatched++
}
func (m *Manager) RecentFiles() []FileRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.files)
	if n > 30 {
		n = 30
	}
	out := make([]FileRecord, n)
	copy(out, m.files[len(m.files)-n:])
	return out
}
func (m *Manager) State() State { m.mu.Lock(); defer m.mu.Unlock(); return m.state }
