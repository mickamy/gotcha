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

	// stop is closed when Watch returns (any exit path),
	// ensuring all goroutines are cleaned up.
	stop := make(chan struct{})
	defer func() {
		close(stop)
		_ = w.Close()
		if err := raw.Exit(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}()

	if err := walkDirs(w, cfg); err != nil {
		return err
	}

	fmt.Println("Watching for changes... (press 'r' to re-run, 'q' to quit)")

	keys := make(chan term.KeyEvent, 1)
	go term.Listen([]string{"r", "R", "q", "Q"}, keys, stop)

	runWithRaw := func() {
		if err := raw.Exit(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
		clearScreen()
		onChange()
		if err := raw.Enter(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	// All test executions are serialized through the trigger channel
	// to prevent concurrent runs and raw mode races.
	trigger := make(chan struct{}, 1)
	go runLoop(trigger, stop, runWithRaw)

	// Initial run.
	trigger <- struct{}{}

	debounce := make(chan struct{}, 1)
	go debounceLoop(debounce, stop, trigger)

	return eventLoop(w, keys, debounce, trigger, stop)
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

func runLoop(trigger <-chan struct{}, stop <-chan struct{}, fn func()) {
	for {
		select {
		case <-trigger:
			fn()
		case <-stop:
			return
		}
	}
}

func debounceLoop(signal <-chan struct{}, stop <-chan struct{}, trigger chan<- struct{}) {
	for {
		// Wait for the first signal.
		select {
		case <-signal:
		case <-stop:
			return
		}

		// Reset the timer on each subsequent signal until quiet.
		timer := time.NewTimer(debounceDelay)
	drain:
		for {
			select {
			case <-signal:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(debounceDelay)
			case <-timer.C:
				select {
				case trigger <- struct{}{}:
				default:
				}
				break drain
			case <-stop:
				timer.Stop()
				return
			}
		}
	}
}

func eventLoop(
	w *fsnotify.Watcher,
	keys <-chan term.KeyEvent,
	debounce chan<- struct{},
	trigger chan<- struct{},
	stop <-chan struct{},
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
				select {
				case trigger <- struct{}{}:
				default:
				}
			}

		case <-stop:
			return nil
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
