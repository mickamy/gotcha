package run

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mickamy/gotcha/internal/config"
)

var (
	summary bool
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
		if summary {
			return RunSummary(cfg)
		}
		return Run(cfg)
	},
}

func init() {
	Cmd.Flags().BoolVarP(&summary, "summary", "s", false, "Output in JSON format")
}

func Run(cfg config.Config) error {
	pkgs, err := cfg.PackagesToTest()
	if err != nil {
		return err
	}

	args := append([]string{"test"}, append(pkgs, cfg.Args...)...)
	fmt.Println("📦 Running: go test")
	for _, pkg := range pkgs {
		fmt.Printf("  %s\n", pkg)
	}
	fmt.Printf("  %s\n\n", strings.Join(cfg.Args, " "))

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

func RunSummary(cfg config.Config) error {
	pkgs, err := cfg.PackagesToTest()
	if err != nil {
		return err
	}

	args := append([]string{"test", "-json"}, append(pkgs, cfg.Args...)...)
	cmd := exec.Command("go", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start test: %w", err)
	}

	decoder := json.NewDecoder(stdout)

	var (
		total       int
		passed      int
		failed      int
		skipped     int
		failedTests []string
	)

	for decoder.More() {
		var evt struct {
			Action  string
			Test    string
			Package string
		}
		if err := decoder.Decode(&evt); err != nil {
			return fmt.Errorf("failed to decode test output: %w", err)
		}

		switch evt.Action {
		case "pass":
			if evt.Test != "" {
				passed++
				total++
			}
		case "fail":
			if evt.Test != "" {
				failed++
				total++
				failedTests = append(failedTests, fmt.Sprintf("%s/%s", evt.Package, evt.Test))
			}
		case "skip":
			if evt.Test != "" {
				skipped++
				total++
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("\033[31m❌ %d failed, %d passed, %d skipped (%d total)\033[0m\n", failed, passed, skipped, total)

		if len(failedTests) > 0 {
			fmt.Println("\n🧨 Failed tests:")
			for _, name := range failedTests {
				fmt.Printf("  - %s\n", name)
			}
		}
		os.Exit(1)
	}

	fmt.Printf("\033[32m✅ %d passed, %d skipped (%d total)\033[0m\n", passed, skipped, total)
	return nil
}
