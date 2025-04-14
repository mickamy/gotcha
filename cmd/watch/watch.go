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
		return Run(cfg)
	},
}

func Run(cfg config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
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

	fmt.Println("👀 Watching for changes...")

	trigger := make(chan struct{}, 1)

	go func() {
		var lastRun time.Time
		for range trigger {
			if time.Since(lastRun) < 300*time.Millisecond {
				continue
			}
			lastRun = time.Now()

			fmt.Print("\033[H\033[2J")
			if err := run.Run(cfg); err != nil {
				fmt.Println(err)
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
				trigger <- struct{}{}
			}
		case err := <-watcher.Errors:
			return err
		}
	}
}
