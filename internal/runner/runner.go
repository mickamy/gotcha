package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Event represents a single go test -json event.
type Event struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output"`
	Elapsed float64   `json:"Elapsed"`
}

// TestID identifies a single test.
type TestID struct {
	Package string
	Test    string
}

func (t TestID) String() string {
	return t.Package + "/" + t.Test
}

// Result holds aggregated test results.
type Result struct {
	Total       int
	Passed      int
	Failed      int
	Skipped     int
	FailedTests []TestID
	Output      map[TestID][]string // per-test output lines
	Duration    time.Duration
	OK          bool
}

// Run executes go test with streaming output.
// Returns true if tests passed.
func Run(ctx context.Context, pkgs, args []string, stdout, stderr io.Writer) (time.Duration, bool) {
	cmdArgs := append([]string{"test"}, append(pkgs, args...)...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	return elapsed, err == nil
}

// RunJSON executes go test -json and returns parsed results.
func RunJSON(ctx context.Context, pkgs, args []string, stderr io.Writer) (Result, error) {
	cmdArgs := append([]string{"test", "-json"}, append(pkgs, args...)...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Stderr = stderr

	var buf bytes.Buffer
	cmd.Stdout = &buf

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	result, err := ParseEvents(&buf)
	result.Duration = elapsed
	if err != nil {
		return result, err
	}
	if runErr != nil && result.Failed == 0 {
		return result, fmt.Errorf("go test: %w", runErr)
	}

	return result, nil
}

// ParseEvents reads go test -json output and aggregates results.
func ParseEvents(r io.Reader) (Result, error) {
	var result Result
	result.Output = make(map[TestID][]string)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var evt Event
		if err := json.Unmarshal(line, &evt); err != nil {
			continue
		}

		if evt.Test == "" {
			continue
		}

		id := TestID{Package: evt.Package, Test: evt.Test}

		switch evt.Action {
		case "output":
			result.Output[id] = append(result.Output[id], evt.Output)
		case "pass":
			result.Total++
			result.Passed++
		case "fail":
			result.Total++
			result.Failed++
			result.FailedTests = append(result.FailedTests, id)
		case "skip":
			result.Total++
			result.Skipped++
		}
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("scan test output: %w", err)
	}

	result.OK = result.Failed == 0

	return result, nil
}

// ListPackages resolves Go package patterns via go list.
func ListPackages(ctx context.Context, patterns []string) ([]string, error) {
	args := append([]string{"list"}, patterns...)
	cmd := exec.CommandContext(ctx, "go", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	var pkgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			pkgs = append(pkgs, line)
		}
	}

	return pkgs, nil
}

// ModulePath returns the current module path.
func ModulePath(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list -m: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
