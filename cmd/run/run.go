package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mickamy/gotcha/internal/config"
)

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run go test once with gotcha config",
	Long:  "Run go test once with gotcha config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return Run(cfg)
	},
}

func Run(cfg config.Config) error {
	args := append([]string{"test"}, append(cfg.Include, cfg.TestFlags...)...)
	fmt.Printf("📦 Running: go %s\n", strings.Join(args, " "))

	cmdExec := exec.Command("go", args...)
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr
	cmdExec.Stdin = os.Stdin

	if err := cmdExec.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	return nil
}
