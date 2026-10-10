package main

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"name-it.vino/internal/config"
	"name-it.vino/internal/history"
	"name-it.vino/internal/pattern"
	"name-it.vino/internal/reminder"
	"name-it.vino/internal/scorer"
	"name-it.vino/internal/session"
	"name-it.vino/internal/updater"
	"name-it.vino/internal/version"
	"name-it.vino/internal/watcher"
)

// App is the object bound to the Wails runtime. Every exported method is
// callable from the frontend through window.go.main.App.
type App struct {
	ctx     context.Context
	session *session.Manager
	watcher *watcher.Watcher
	history *history.Store

	mu            sync.RWMutex
	config        config.Config
	noticeLevel   int
	noticeActive  bool
	pendingUpdate *updater.Release
}

func NewApp() *App {
	return &App{
		session: session.New(),
		history: history.New(),
		config:  config.Default(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// The update check runs off the UI thread so a slow network can never delay
	// the window from appearing.
	go a.checkForUpdates()
}

// checkForUpdates quietly asks GitHub for the latest release and, when a newer
// one exists, stores it for the UI. Failures are deliberately swallowed: being
// offline is not an error worth bothering the user with.
func (a *App) checkForUpdates() {
	release, err := updater.Check(version.Version)
	if err != nil || release == nil {
		return
	}

	a.mu.Lock()
	a.pendingUpdate = release
	a.mu.Unlock()

	if a.ctx == nil {
		return
	}

	// The frontend may not have subscribed yet; GetPendingUpdate covers that gap.
	runtime.EventsEmit(a.ctx, "nameit:update", release)
}

func (a *App) ChooseFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Choose your battlefield"})
}

func (a *App) StartWorking(root string) error {
	if a.watcher != nil {
		a.watcher.Stop()
	}

	a.session.Start(root)
	a.watcher = watcher.New(root, a.handleEvent)

	return a.watcher.Start()
}

func (a *App) StopWorking() {
	if a.watcher != nil {
		a.watcher.Stop()
		a.watcher = nil
	}

	a.session.Stop()
}

func (a *App) GetState() session.State    { return a.session.State() }
func (a *App) GetCrimes() []history.Crime { return a.history.List(40) }
func (a *App) GetVersion() string         { return version.Display() }

func (a *App) GetConfig() config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.config
}

func (a *App) SetAggressiveness(level string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.Aggressiveness = level
}

func (a *App) Pause(minutes int) {
	a.session.Pause(time.Duration(minutes) * time.Minute)
}

// GetPendingUpdate returns a newer release discovered at startup, or nil when
// the app is up to date.
func (a *App) GetPendingUpdate() *updater.Release {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.pendingUpdate
}

// OpenUpdatePage sends the user to the release page of the pending update.
func (a *App) OpenUpdatePage() {
	a.mu.RLock()
	release := a.pendingUpdate
	a.mu.RUnlock()

	if release == nil || release.URL == "" || a.ctx == nil {
		return
	}

	runtime.BrowserOpenURL(a.ctx, release.URL)
}

// handleEvent inspects a newly created file, scores its name and, when it
// deserves it, escalates into a full-window interruption.
func (a *App) handleEvent(ev watcher.FileEvent) {
	if ev.Type != watcher.Create {
		return
	}

	a.mu.RLock()
	cfg := a.config
	a.mu.RUnlock()

	filename := filepath.Base(ev.Path)

	if !matchesNamingRule(cfg, filename) {
		a.session.Record(ev.Path)
		return
	}

	assessment := scorer.Assess(ev.Path, a.session.RecentFiles())
	a.session.Record(ev.Path)

	a.history.Add(history.Crime{
		Timestamp: time.Now(),
		Filename:  filename,
		Score:     assessment.Score,
		Severity:  assessment.Severity,
		Action:    "notified",
	})

	if a.ctx == nil || a.session.Paused() {
		return
	}

	// Each serious crime nudges the annoyance level up; clean names slowly
	// bring it back down. The level is capped so it can plateau at maximum.
	a.mu.Lock()
	if assessment.Score >= 30 {
		a.noticeLevel++
		if a.noticeLevel > 6 {
			a.noticeLevel = 6
		}
	} else if a.noticeLevel > 0 {
		a.noticeLevel--
	}
	level := a.noticeLevel
	a.noticeActive = true
	a.mu.Unlock()

	message := reminder.MessageForLevel(assessment.Severity, level, filename)

	runtime.EventsEmit(a.ctx, "nameit:crime", map[string]interface{}{
		"filename":   filename,
		"score":      assessment.Score,
		"confidence": assessment.Confidence,
		"severity":   assessment.Severity,
		"message":    message,
		"reasons":    assessment.Reasons,
	})

	_ = beeep.Notify("NAME IT", message, "")

	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowFullscreen(a.ctx)

	// Some window managers ignore the first fullscreen request when the window
	// is still minimised, so the raise is repeated just after the animation.
	time.AfterFunc(150*time.Millisecond, func() {
		a.mu.RLock()
		active := a.noticeActive
		a.mu.RUnlock()

		if a.ctx == nil || !active {
			return
		}

		runtime.WindowShow(a.ctx)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
	})
}

func (a *App) DismissNotice() {
	a.mu.Lock()
	if !a.noticeActive {
		a.mu.Unlock()
		return
	}
	a.noticeActive = false
	a.mu.Unlock()

	runtime.WindowSetAlwaysOnTop(a.ctx, false)

	if runtime.WindowIsFullscreen(a.ctx) {
		runtime.WindowUnfullscreen(a.ctx)
	}

	runtime.WindowMinimise(a.ctx)
}

func (a *App) SetNamingMode(mode string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.NamingMode = mode
}

func (a *App) SetCustomPattern(custom string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.CustomPattern = custom
}

// matchesNamingRule reports whether a filename should be judged at all. A file
// that already follows the selected convention is left alone.
func matchesNamingRule(cfg config.Config, filename string) bool {
	switch cfg.NamingMode {

	case "pattern_1":
		return !pattern.Matches(filename, pattern.DefaultPattern())

	case "pattern_2":
		return !pattern.Matches(filename, pattern.DatePattern())

	case "custom":
		if cfg.CustomPattern == "" {
			return true
		}
		return !pattern.Matches(filename, cfg.CustomPattern)

	case "every_file":
		return true

	default:
		return true
	}
}
