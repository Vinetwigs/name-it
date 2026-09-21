package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"name-it.vino/internal/config"
	"name-it.vino/internal/history"
	"name-it.vino/internal/pattern"
	"name-it.vino/internal/reminder"
	"name-it.vino/internal/scorer"
	"name-it.vino/internal/session"
	"name-it.vino/internal/watcher"
)

type App struct {
	ctx         context.Context
	session     *session.Manager
	watcher     *watcher.Watcher
	history     *history.Store
	noticeLevel int
	config      config.Config

	noticeActive bool
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

func (a *App) GetState() session.State        { return a.session.State() }
func (a *App) GetCrimes() []history.Crime     { return a.history.List(40) }
func (a *App) GetConfig() config.Config       { return a.config }
func (a *App) SetAggressiveness(level string) { a.config.Aggressiveness = level }
func (a *App) Pause(minutes int)              { a.session.Pause(time.Duration(minutes) * time.Minute) }

func (a *App) handleEvent(ev watcher.FileEvent) {

	if ev.Type != watcher.Create {
		return
	}

	filename := filepath.Base(ev.Path)

	if !a.shouldNotify(filename) {
		a.session.Record(ev.Path)
		return
	}

	assessment := scorer.Assess(ev.Path, a.session.RecentFiles())
	a.session.Record(ev.Path)
	a.history.Add(history.Crime{
		Timestamp: time.Now(),
		Filename:  filepath.Base(ev.Path),
		Score:     assessment.Score,
		Severity:  assessment.Severity,
		Action:    "notified",
	})

	if a.ctx == nil || a.session.Paused() {
		return
	}

	if assessment.Score >= 30 {
		a.noticeLevel++
		if a.noticeLevel > 6 {
			a.noticeLevel = 6
		}
	} else if a.noticeLevel > 0 {
		a.noticeLevel--
	}
	msg := reminder.MessageForLevel(assessment.Severity, a.noticeLevel, filepath.Base(ev.Path))
	runtime.EventsEmit(a.ctx, "nameit:crime", map[string]interface{}{
		"filename":   filepath.Base(ev.Path),
		"score":      assessment.Score,
		"confidence": assessment.Confidence,
		"severity":   assessment.Severity,
		"message":    msg,
		"reasons":    assessment.Reasons,
	})

	_ = beeep.Notify("NAME IT", msg, "")

	a.noticeActive = true

	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowFullscreen(a.ctx)

	time.AfterFunc(150*time.Millisecond, func() {
		if a.ctx == nil || !a.noticeActive {
			return
		}

		runtime.WindowShow(a.ctx)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
	})
}

func (a *App) DismissNotice() {
	if !a.noticeActive {
		return
	}

	a.noticeActive = false

	runtime.WindowSetAlwaysOnTop(a.ctx, false)

	if runtime.WindowIsFullscreen(a.ctx) {
		runtime.WindowUnfullscreen(a.ctx)
	}

	runtime.WindowMinimise(a.ctx)
}

func (a *App) SetNamingMode(mode string) {
	a.config.NamingMode = mode
}

func (a *App) SetCustomPattern(pattern string) {
	a.config.CustomPattern = pattern
}

func (a *App) shouldNotify(filename string) bool {
	switch a.config.NamingMode {

	case "every_file":
		return true

	case "pattern_1":
		return !pattern.Matches(filename, "filename_vX.extension")

	case "pattern_2":
		return !pattern.Matches(filename, "YYYY-MM-DD_filename_vX.extension")

	case "custom":
		if a.config.CustomPattern == "" {
			return true
		}

		return !pattern.Matches(
			filename,
			a.config.CustomPattern,
		)

	default:
		return true
	}
}
