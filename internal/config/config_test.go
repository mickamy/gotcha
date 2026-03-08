package config_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mickamy/gotcha/internal/config"
)

func TestLoadByPath(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadByPath("./testdata/.gotcha.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantInclude := []string{"./..."}
	wantExclude := []string{"vendor/", "mocks/"}
	wantArgs := []string{"-v"}

	if !slices.Equal(cfg.Include, wantInclude) {
		t.Errorf("Include: got %v, want %v", cfg.Include, wantInclude)
	}
	if !slices.Equal(cfg.Exclude, wantExclude) {
		t.Errorf("Exclude: got %v, want %v", cfg.Exclude, wantExclude)
	}
	if !slices.Equal(cfg.Args, wantArgs) {
		t.Errorf("Args: got %v, want %v", cfg.Args, wantArgs)
	}
}

func TestLoadByPath_NotFound(t *testing.T) {
	t.Parallel()

	_, err := config.LoadByPath("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadByPath_InvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".gotcha.yaml")
	if err := os.WriteFile(path, []byte(":\n  :\n---\n: ["), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.LoadByPath(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	if len(cfg.Include) == 0 {
		t.Fatal("default Include should not be empty")
	}
	if len(cfg.Exclude) == 0 {
		t.Fatal("default Exclude should not be empty")
	}
	if len(cfg.Args) == 0 {
		t.Fatal("default Args should not be empty")
	}
}

func TestSave_And_Load(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".gotcha.yaml")

	original := config.Config{
		Include: []string{"./cmd/...", "./internal/..."},
		Exclude: []string{"testdata/"},
		Args:    []string{"-v", "-count=1"},
	}

	if err := config.Save(path, original); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := config.LoadByPath(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !slices.Equal(loaded.Include, original.Include) {
		t.Errorf("Include: got %v, want %v", loaded.Include, original.Include)
	}
	if !slices.Equal(loaded.Exclude, original.Exclude) {
		t.Errorf("Exclude: got %v, want %v", loaded.Exclude, original.Exclude)
	}
	if !slices.Equal(loaded.Args, original.Args) {
		t.Errorf("Args: got %v, want %v", loaded.Args, original.Args)
	}
}

func TestShouldExclude(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Exclude: []string{"vendor/", "mocks/"},
	}

	tests := []struct {
		path string
		want bool
	}{
		{"github.com/foo/bar/vendor/pkg", true},
		{"github.com/foo/bar/mocks/mock_repo.go", true},
		{"github.com/foo/bar/internal/service", false},
		{"vendor/something", true},
		{"internal/mocks/repo", true},
		{"cmd/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			if got := cfg.ShouldExclude(tt.path); got != tt.want {
				t.Errorf("ShouldExclude(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

