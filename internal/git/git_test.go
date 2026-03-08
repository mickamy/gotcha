package git_test

import (
	"slices"
	"testing"

	"github.com/mickamy/gotcha/internal/git"
)

func TestExtractPackages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  []string
	}{
		{
			name:  "single go file",
			input: []byte("internal/config/config.go\n"),
			want:  []string{"./internal/config"},
		},
		{
			name:  "multiple files same package",
			input: []byte("internal/config/config.go\ninternal/config/config_test.go\n"),
			want:  []string{"./internal/config"},
		},
		{
			name: "multiple packages",
			input: []byte(
				"internal/config/config.go\n" +
					"internal/runner/runner.go\n" +
					"main.go\n",
			),
			want: []string{"./internal/config", "./internal/runner", "."},
		},
		{
			name:  "non-go files ignored",
			input: []byte("README.md\ngo.mod\ninternal/config/config.go\n"),
			want:  []string{"./internal/config"},
		},
		{
			name:  "empty input",
			input: []byte(""),
			want:  nil,
		},
		{
			name:  "only non-go files",
			input: []byte("README.md\nMakefile\n"),
			want:  nil,
		},
		{
			name:  "root go file",
			input: []byte("main.go\n"),
			want:  []string{"."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := git.ExtractPackages(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
