package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangedPackages returns Go package paths (relative, e.g. "./internal/config")
// that contain files changed since the given git ref (e.g. "HEAD", "main").
// Only .go files are considered.
func ChangedPackages(ref string) ([]string, error) {
	out, err := gitDiff(ref)
	if err != nil {
		return nil, err
	}

	return extractPackages(out), nil
}

// ChangedPackagesUncommitted returns Go package paths with uncommitted changes
// (staged, unstaged, and untracked files).
func ChangedPackagesUncommitted() ([]string, error) {
	sources := []func() ([]byte, error){
		gitDiffStaged,
		gitDiffUnstaged,
		gitUntracked,
	}

	seen := make(map[string]struct{})
	var pkgs []string
	for _, fn := range sources {
		out, err := fn()
		if err != nil {
			return nil, err
		}
		for _, p := range extractPackages(out) {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			pkgs = append(pkgs, p)
		}
	}

	return pkgs, nil
}

func gitDiff(ref string) ([]byte, error) {
	cmd := exec.Command("git", "diff", "--name-only", ref)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only %s: %w", ref, err)
	}
	return out, nil
}

func gitDiffStaged() ([]byte, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached: %w", err)
	}
	return out, nil
}

func gitDiffUnstaged() ([]byte, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}

func gitUntracked() ([]byte, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files --others: %w", err)
	}
	return out, nil
}

func extractPackages(out []byte) []string {
	seen := make(map[string]struct{})
	var pkgs []string

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file == "" {
			continue
		}
		if filepath.Ext(file) != ".go" {
			continue
		}

		dir := filepath.Dir(file)
		pkg := "./" + dir
		if dir == "." {
			pkg = "."
		}

		if _, ok := seen[pkg]; ok {
			continue
		}
		seen[pkg] = struct{}{}
		pkgs = append(pkgs, pkg)
	}

	return pkgs
}
