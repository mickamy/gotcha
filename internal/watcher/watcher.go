package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/mickamy/gotcha/internal/config"
	"github.com/mickamy/gotcha/internal/term"
)

const debounceDelay = 500 * time.Millisecond

// OnChange is called when file changes are detected or a manual rerun is triggered.
type OnChange func()

// Watch monitors .go file changes and keyboard input, calling onChange as needed.
// It blocks until quit is signaled (q/Q/ctrl+c/ctrl+d).
func Watch(cfg config.Config, onChange OnChange) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	raw := term.NewRawMode(int(os.Stdin.Fd()))
	defer func() {
		_ = w.Close()
		_ = raw.Exit()
	}()

	if err := walkDirs(w, cfg); err != nil {
		return err
	}

	fmt.Println("Watching for changes... (press 'r' to re-run, 'q' to quit)")

	keys := make(chan term.KeyEvent, 1)
	done := make(chan struct{})

	go func() {
		term.Listen([]string{"r", "R", "q", "Q"}, keys, done)
		close(done)
	}()

	runWithRaw := func() {
		_ = raw.Exit()
		clearScreen()
		onChange()
		_ = raw.Enter()
	}

	runWithRaw()

	debounce := make(chan struct{}, 1)
	go debounceLoop(debounce, done, runWithRaw)

	return eventLoop(w, keys, debounce, done, runWithRaw)
}

func walkDirs(w *fsnotify.Watcher, cfg config.Config) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if cfg.ShouldExclude(path) {
			return filepath.SkipDir
		}
		return w.Add(path)
	})
}

func debounceLoop(signal <-chan struct{}, done <-chan struct{}, fn func()) {
	for {
		select {
		case <-signal:
		case <-done:
			return
		}

		timer := time.NewTimer(debounceDelay)
		select {
		case <-timer.C:
			fn()
			// Drain signals that arrived during fn() to prevent double-run.
			// Fixes: https://github.com/mickamy/gotcha/issues/12
			drain(signal)
		case <-done:
			timer.Stop()
			return
		}
	}
}

func drain(ch <-chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func eventLoop(
	w *fsnotify.Watcher,
	keys <-chan term.KeyEvent,
	debounce chan<- struct{},
	done <-chan struct{},
	onRerun func(),
) error {
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if filepath.Ext(event.Name) == ".go" {
				select {
				case debounce <- struct{}{}:
				default:
				}
			}

		case err := <-w.Errors:
			return fmt.Errorf("watcher error: %w", err)

		case key := <-keys:
			switch {
			case key.Key == "q" || key.Key == "Q" || key.Key == "ctrl+c" || key.EOF:
				fmt.Println("\nExiting...")
				return nil
			case key.Key == "r" || key.Key == "R":
				onRerun()
			}

		case <-done:
			return nil
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
