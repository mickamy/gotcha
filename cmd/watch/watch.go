package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/mickamy/gotcha/cmd/run"
	"github.com/mickamy/gotcha/internal/config"
	"github.com/mickamy/gotcha/internal/stdin"
)

var (
	summary bool
)

var Cmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and run tests automatically",
	Long:  "Watch for file changes and run tests automatically",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return Run(cfg, summary)
	},
}

func init() {
	Cmd.Flags().BoolVarP(&summary, "summary", "s", false, "Summarize test results")
}

func Run(cfg config.Config, summary bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
		_ = stdin.ExitRawMode()
	}(watcher)

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if cfg.ShouldExclude(path) {
			fmt.Println("Skipping excluded path:", path)
			return nil
		}
		return watcher.Add(path)
	})
	if err != nil {
		return err
	}

	fmt.Println("👀 Watching for changes... (press 'r' to re-run, 'q' or 'ctrl+d' to quit)")

	keys := make(chan stdin.KeyPressDownEvent, 1)
	done := make(chan struct{})

	go func() {
		if err := stdin.Listen([]string{"r", "R", "q", "Q"}, keys); err != nil {
			fmt.Println("⚠️ Failed to listen for keys:", err)
		}
		close(done)
	}()

	runTests := func() {
		_ = stdin.ExitRawMode()
		fmt.Print("\033[H\033[2J")
		if summary {
			run.RunSummary(cfg, true)
		} else {
			run.Run(cfg, true)
		}
		_ = stdin.EnterRawMode()
	}

	runTests()

	debounceDelay := 1000 * time.Millisecond
	debounceSignal := make(chan struct{}, 1)

	go func() {
		var timer *time.Timer
		for {
			<-debounceSignal
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(debounceDelay)

			select {
			case <-timer.C:
				runTests()
			case <-done:
				timer.Stop()
				return
			}
		}
	}()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Ext(event.Name) == ".go" {
				select {
				case debounceSignal <- struct{}{}:
				default:
				}
			}

		case err := <-watcher.Errors:
			return err

		case key := <-keys:
			switch key.Key {
			case "r", "R":
				runTests()
			case "q", "Q", "ctrl+c":
				fmt.Println("\n👋 Exiting...")
				return nil
			}
			if key.EOF {
				fmt.Println("\n👋 Received EOF (ctrl+d), exiting...")
				return nil
			}

		case <-done:
			return nil
		}
	}
}
