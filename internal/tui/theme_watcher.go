package tui

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Pi's theme watcher (theme.js startThemeWatcher): while a custom theme is
// active, edits to its file apply live, 100 ms after the last change. A file
// that is missing or invalid while being edited keeps the last good theme.

const themeReloadDelay = 100 * time.Millisecond

type themeWatcher struct {
	watcher  *fsnotify.Watcher
	done     chan struct{}
	exited   chan struct{}
	stopOnce sync.Once
}

func (w *themeWatcher) stop() {
	if w != nil {
		w.stopOnce.Do(func() {
			close(w.done)
			_ = w.watcher.Close()
		})
		<-w.exited
	}
}

// customThemePath is the file a custom theme name loads from, or "".
func customThemePath(name string) string {
	for _, dir := range customThemeDirs() {
		path := filepath.Join(dir, name+".json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// watchActiveTheme watches the active theme's file when it is a custom
// theme, replacing any previous watcher.
func (c *chatTUI) watchActiveTheme() {
	c.themeWatcher.stop()
	c.themeWatcher = nil
	name := piActiveTheme
	if _, builtin := piBuiltinThemes[name]; builtin || name == systemThemeName {
		return
	}
	path := customThemePath(name)
	if path == "" {
		return
	}
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	if err := fw.Add(filepath.Dir(path)); err != nil {
		_ = fw.Close()
		return
	}
	w := &themeWatcher{watcher: fw, done: make(chan struct{}), exited: make(chan struct{})}
	c.themeWatcher = w
	go func() {
		defer close(w.exited)
		var timer *time.Timer
		var tick <-chan time.Time
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()
		reload := func() {
			c.runOnUI(func() {
				select {
				case <-w.done:
					return
				default:
				}
				if c.themeWatcher != w || piActiveTheme != name {
					return // a stale timer after a theme switch
				}
				if _, err := os.Stat(path); err != nil {
					return
				}
				colors, err := loadCustomTheme(name)
				if err != nil {
					return
				}
				applyCustomThemeColors(name, colors)
				c.invalidateFooter()
				c.markDirty()
			})
		}
		for {
			select {
			case <-w.done:
				return
			case <-tick:
				tick = nil
				reload()
			case ev, ok := <-fw.Events:
				if !ok {
					return
				}
				if filepath.Base(ev.Name) != name+".json" {
					continue
				}
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(themeReloadDelay)
				tick = timer.C
			case _, ok := <-fw.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}
