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

	// Establish the initial baseline.
	//
	// Files that already exist when monitoring starts are not
	// considered new files.
	w.fileCount = w.countFiles()

	w.wg.Add(1)

	go w.loop()

	return nil
}

func (w *Watcher) addTree(root string) error {
	return filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() &&
				!strings.HasPrefix(info.Name(), ".") {
				return w.fw.Add(path)
			}

			return nil
		},
	)
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
			// Filesystem watcher errors are intentionally ignored here.
			// The watcher will continue processing future events.
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {

	// NAME IT only cares about CREATE events.
	//
	// Rename, Remove and Write are ignored here.
	if event.Op&fsnotify.Create == 0 {
		return
	}

	info, err := os.Stat(event.Name)
	if err != nil {
		return
	}

	// A newly created directory is not a file crime.
	// Start watching it instead.
	if info.IsDir() {
		_ = w.fw.Add(event.Name)
		return
	}

	if w.shouldIgnore(event.Name) {
		return
	}

	// Wait a moment before deciding whether this CREATE actually
	// increased the number of files.
	//
	// This is important on Windows because a rename can sometimes
	// appear as a REMOVE + CREATE sequence.
	go w.checkNewFile(event.Name)
}

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

	// Always update the baseline after the filesystem settles.
	w.fileCount = currentCount

	w.mu.Unlock()

	// The total number of files did not increase.
	//
	// This covers rename-like REMOVE + CREATE sequences.
	if currentCount <= previousCount {
		return
	}

	w.emit(path, Create)
}

func (w *Watcher) countFiles() int {
	count := 0

	_ = filepath.Walk(
		w.root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if path == w.root {
				return nil
			}

			if info.IsDir() {
				// Do not count hidden directories.
				//
				// This also keeps the counting logic consistent
				// with addTree().
				if strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}

				return nil
			}

			// Only actual files are counted.
			if info.Mode().IsRegular() {
				count++
			}

			return nil
		},
	)

	return count
}

func (w *Watcher) shouldIgnore(path string) bool {
	name := filepath.Base(path)

	if strings.HasPrefix(name, ".~lock") {
		return true
	}

	if strings.Contains(
		strings.ToLower(name),
		"autosave",
	) {
		return true
	}

	return false
}

func (w *Watcher) emit(path string, eventType EventType) {
	now := time.Now()

	w.mu.Lock()

	last, exists := w.seen[path]

	if exists &&
		now.Sub(last) < 250*time.Millisecond {
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
