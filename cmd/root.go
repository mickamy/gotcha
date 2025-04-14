package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mickamy/gotcha/cmd/version"
)

var (
	versionFlag bool
)

var cmd = &cobra.Command{
	Use:   "gotcha",
	Short: "A Go test watcher",
	Long:  `gotcha is a CLI tool to automatically run go test on file changes.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("gotcha version %s\n", "dev")
			os.Exit(0)
		}
	},
}

func init() {
	cmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Show version and exit")
	cmd.AddCommand(version.Cmd)
}

func Execute() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
