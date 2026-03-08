package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mickamy/gotcha/internal/config"
	"github.com/mickamy/gotcha/internal/git"
	"github.com/mickamy/gotcha/internal/runner"
	"github.com/mickamy/gotcha/internal/watcher"
)

var version = "dev"

const usage = `gotcha - Go test watcher

Usage:
  gotcha <command> [options]

Commands:
  run      Run tests once
  watch    Watch for file changes and run tests
  init     Generate .gotcha.yaml
  version  Print version

Run 'gotcha <command> -h' for command-specific help.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	subcmd := os.Args[1]

	switch subcmd {
	case "run":
		runCmd(os.Args[2:])
	case "watch":
		watchCmd(os.Args[2:])
	case "init":
		initCmd()
	case "version":
		fmt.Println("gotcha", version)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcmd)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func runCmd(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fast := fs.Bool("fast", false, "Only test packages with uncommitted changes")
	focus := fs.Bool("focus", false, "Show only failed test output")
	summary := fs.Bool("summary", false, "Show test summary (pass/fail counts)")
	_ = fs.Parse(args)

	if *focus && *summary {
		fmt.Fprintln(os.Stderr, "error: --focus and --summary are mutually exclusive")
		os.Exit(1)
	}

	cfg := loadConfig()
	ctx := context.Background()
	pkgs := resolvePackages(ctx, cfg, *fast)

	if len(pkgs) == 0 {
		fmt.Println("No packages to test.")
		return
	}

	if *focus || *summary {
		ok := runJSON(ctx, pkgs, cfg.Args, *focus)
		if !ok {
			os.Exit(1)
		}
		return
	}

	printRunHeader(pkgs, cfg.Args)
	elapsed, ok := runner.Run(ctx, pkgs, cfg.Args, os.Stdout, os.Stderr)
	printResult(os.Stdout, elapsed, ok)
	if !ok {
		os.Exit(1)
	}
}

func watchCmd(args []string) {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	fast := fs.Bool("fast", false, "Only test packages with uncommitted changes")
	focus := fs.Bool("focus", false, "Show only failed test output")
	summary := fs.Bool("summary", false, "Show test summary (pass/fail counts)")
	_ = fs.Parse(args)

	if *focus && *summary {
		fmt.Fprintln(os.Stderr, "error: --focus and --summary are mutually exclusive")
		os.Exit(1)
	}

	cfg := loadConfig()

	onChange := func() {
		ctx := context.Background()
		pkgs := resolvePackages(ctx, cfg, *fast)

		if len(pkgs) == 0 {
			fmt.Println("No packages to test.")
			return
		}

		if *focus || *summary {
			runJSON(ctx, pkgs, cfg.Args, *focus)
			return
		}

		printRunHeader(pkgs, cfg.Args)
		elapsed, ok := runner.Run(ctx, pkgs, cfg.Args, os.Stdout, os.Stderr)
		printResult(os.Stdout, elapsed, ok)
	}

	if err := watcher.Watch(cfg, onChange); err != nil {
		fatal(err)
	}
}

func initCmd() {
	if _, err := os.Stat(config.DefaultPath); err == nil {
		fmt.Print(".gotcha.yaml already exists. Overwrite? [y/N]: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		res := strings.TrimSpace(scanner.Text())
		if res != "y" && res != "Y" {
			fmt.Println("Canceled.")
			return
		}
	}

	if err := config.Save(config.DefaultPath, config.Default()); err != nil {
		fatal(err)
	}
	fmt.Println(".gotcha.yaml generated.")
}

func loadConfig() config.Config {
	cfg, err := config.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return config.Default()
		}
		fatal(err)
	}
	return cfg
}

func resolvePackages(ctx context.Context, cfg config.Config, fast bool) []string {
	if fast {
		pkgs, err := git.ChangedPackagesUncommitted()
		if err != nil {
			fatal(err)
		}
		return filterExcluded(cfg, pkgs)
	}

	pkgs, err := runner.ListPackages(ctx, cfg.Include)
	if err != nil {
		fatal(err)
	}

	mod, err := runner.ModulePath(ctx)
	if err != nil {
		fatal(err)
	}

	var resolved []string
	for _, pkg := range pkgs {
		rel := pkg
		if pkg == mod || strings.HasPrefix(pkg, mod+"/") {
			trimmed := strings.TrimPrefix(pkg, mod)
			if trimmed == "" {
				rel = "."
			} else {
				rel = "./" + strings.TrimPrefix(trimmed, "/")
			}
		}
		resolved = append(resolved, rel)
	}

	return filterExcluded(cfg, resolved)
}

func filterExcluded(cfg config.Config, pkgs []string) []string {
	var filtered []string
	for _, pkg := range pkgs {
		if !cfg.ShouldExclude(pkg) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func runJSON(ctx context.Context, pkgs, args []string, focus bool) bool {
	result, err := runner.RunJSON(ctx, pkgs, args, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}

	if focus {
		printFocusOutput(os.Stdout, result)
	} else {
		printSummary(os.Stdout, result)
	}

	return result.OK
}

func printRunHeader(pkgs, args []string) {
	fmt.Println("Running: go test")
	for _, pkg := range pkgs {
		fmt.Printf("  %s\n", pkg)
	}
	if len(args) > 0 {
		fmt.Printf("  %s\n", strings.Join(args, " "))
	}
	fmt.Println()
}

func printResult(w io.Writer, elapsed interface{ String() string }, ok bool) {
	if ok {
		_, _ = fmt.Fprintf(w, "\033[32mAll tests passed (%s)\033[0m\n", elapsed)
	} else {
		_, _ = fmt.Fprintf(w, "\033[31mTests failed (%s)\033[0m\n", elapsed)
	}
}

func printSummary(w io.Writer, result runner.Result) {
	printPackageErrors(w, result)
	if result.OK {
		_, _ = fmt.Fprintf(w,
			"\033[32m%d passed, %d skipped (%d total, %s)\033[0m\n",
			result.Passed, result.Skipped, result.Total, result.Duration,
		)
	} else {
		_, _ = fmt.Fprintf(w,
			"\033[31m%d failed, %d passed, %d skipped (%d total, %s)\033[0m\n",
			result.Failed, result.Passed, result.Skipped, result.Total, result.Duration,
		)
		printFailedTests(w, result)
	}
}

func printFocusOutput(w io.Writer, result runner.Result) {
	printPackageErrors(w, result)
	if result.OK {
		_, _ = fmt.Fprintf(w,
			"\033[32mAll %d tests passed (%s)\033[0m\n",
			result.Total, result.Duration,
		)
		return
	}

	for _, id := range result.FailedTests {
		_, _ = fmt.Fprintf(w, "\033[31m--- FAIL: %s\033[0m\n", id)
		if lines, ok := result.Output[id]; ok {
			for _, line := range lines {
				_, _ = fmt.Fprint(w, "    ", line)
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	_, _ = fmt.Fprintf(w,
		"\033[31m%d failed, %d passed (%d total, %s)\033[0m\n",
		result.Failed, result.Passed, result.Total, result.Duration,
	)
}

func printFailedTests(w io.Writer, result runner.Result) {
	if len(result.FailedTests) == 0 {
		return
	}
	_, _ = fmt.Fprintln(w, "\nFailed tests:")
	for _, id := range result.FailedTests {
		_, _ = fmt.Fprintf(w, "  - %s\n", id)
	}
}

func printPackageErrors(w io.Writer, result runner.Result) {
	for _, pkg := range result.FailedPackages {
		_, _ = fmt.Fprintf(w, "\033[31m--- FAIL: %s (build)\033[0m\n", pkg)
		if lines, ok := result.PackageOutput[pkg]; ok {
			for _, line := range lines {
				_, _ = fmt.Fprint(w, "    ", line)
			}
		}
		_, _ = fmt.Fprintln(w)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
