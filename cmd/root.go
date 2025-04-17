package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	initPkg "github.com/mickamy/gotcha/cmd/init"
	"github.com/mickamy/gotcha/cmd/run"
	"github.com/mickamy/gotcha/cmd/version"
	"github.com/mickamy/gotcha/cmd/watch"
)

var cmd = &cobra.Command{
	Use:   "gotcha",
	Short: "A Go test watcher",
	Long:  `gotcha is a CLI tool to automatically run go test on file changes.`,
}

func init() {
	cmd.AddCommand(initPkg.Cmd)
	cmd.AddCommand(run.Cmd)
	cmd.AddCommand(version.Cmd)
	cmd.AddCommand(watch.Cmd)
}

func Execute() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
