package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	pkgs, err := cfg.PackagesToTest()
	if err != nil {
		return err
	}

	args := append([]string{"test"}, append(pkgs, cfg.Args...)...)
	fmt.Printf("📦 Running: go %s\n", strings.Join(args, " "))

	cmdExec := exec.Command("go", args...)
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr
	cmdExec.Stdin = os.Stdin

	start := time.Now()
	if err := cmdExec.Run(); err != nil {
		fmt.Printf("\033[31m❌ Tests failed (%s)\033[0m\n", time.Since(start))
		return fmt.Errorf("go test failed: %w", err)
	}

	fmt.Printf("\033[32m✅ All tests passed (%s)\033[0m\n", time.Since(start))
	return nil
}
