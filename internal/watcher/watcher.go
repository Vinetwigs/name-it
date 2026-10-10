// Package watcher wraps fsnotify with the handful of rules Name It needs:
// follow new directories, ignore editor noise, and report only files that were
// genuinely created (not rewritten by a rename).
package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type EventType string

const (
	Create EventType = "create"
	Rename EventType = "rename"
	Remove EventType = "remove"
	Write  EventType = "write"
)

type FileEvent struct {
	Path      string
	Type      EventType
	Timestamp time.Time
}

type Watcher struct {
	root string
	fw   *fsnotify.Watcher
	cb   func(FileEvent)

	stop chan struct{}
	wg   sync.WaitGroup

	mu        sync.Mutex
	seen      map[string]time.Time
	fileCount int
}

func New(root string, cb func(FileEvent)) *Watcher {
	return &Watcher{
		root: root,
		cb:   cb,
		stop: make(chan struct{}),
		seen: make(map[string]time.Time),
	}
}

func (w *Watcher) Start() error {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	w.fw = fw

	if err := w.addTree(w.root); err != nil {
		_ = fw.Close()
		return err
	}

	// Files already present at startup form the baseline, so they are not
	// reported as new.
	w.fileCount = w.countFiles()

	w.wg.Add(1)
	go w.loop()

	return nil
}

// addTree registers every visible directory below root. Hidden directories are
// skipped on purpose so the watcher never dives into .git or similar.
func (w *Watcher) addTree(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			return w.fw.Add(path)
		}

		return nil
	})
}

func (w *Watcher) Stop() {
	select {
	case <-w.stop:
		// Already stopped.
	default:
		close(w.stop)
	}

	if w.fw != nil {
		_ = w.fw.Close()
	}

	w.wg.Wait()
}

func (w *Watcher) loop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.stop:
			return

		case event, ok := <-w.fw.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case <-w.fw.Errors:
			// Watcher errors are transient and self-healing; keeping the loop
			// alive matters more than reporting them.
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only creations matter. Renames, removals and writes are not crimes.
	if event.Op&fsnotify.Create == 0 {
		return
	}

	info, err := os.Stat(event.Name)
	if err != nil {
		return
	}

	if info.IsDir() {
		_ = w.fw.Add(event.Name)
		return
	}

	if w.shouldIgnore(event.Name) {
		return
	}

	// Confirm asynchronously: on Windows a rename often arrives as a
	// REMOVE followed by a CREATE, which must not be mistaken for a new file.
	go w.checkNewFile(event.Name)
}

// checkNewFile waits for the filesystem to settle, then reports the file only
// if the total file count actually grew.
func (w *Watcher) checkNewFile(path string) {
	const debounce = 200 * time.Millisecond

	select {
	case <-w.stop:
		return

	case <-time.After(debounce):
	}

	currentCount := w.countFiles()

	w.mu.Lock()
	previousCount := w.fileCount
	w.fileCount = currentCount
	w.mu.Unlock()

	if currentCount <= previousCount {
		return
	}

	w.emit(path, Create)
}

// countFiles returns the number of regular files below root, ignoring hidden
// directories so the count stays consistent with addTree.
func (w *Watcher) countFiles() int {
	count := 0

	_ = filepath.Walk(w.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == w.root {
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() {
			count++
		}

		return nil
	})

	return count
}

// shouldIgnore filters out lock files and autosave drafts, which are created
// constantly while working and say nothing about the user's naming habits.
func (w *Watcher) shouldIgnore(path string) bool {
	name := filepath.Base(path)

	if strings.HasPrefix(name, ".~lock") {
		return true
	}

	if strings.Contains(strings.ToLower(name), "autosave") {
		return true
	}

	return false
}

// emit delivers an event to the callback, collapsing duplicates that arrive
// within a quarter of a second.
func (w *Watcher) emit(path string, eventType EventType) {
	now := time.Now()

	w.mu.Lock()

	last, exists := w.seen[path]
	if exists && now.Sub(last) < 250*time.Millisecond {
		w.mu.Unlock()
		return
	}

	w.seen[path] = now
	w.mu.Unlock()

	w.cb(FileEvent{
		Path:      path,
		Type:      eventType,
		Timestamp: now,
	})
}
