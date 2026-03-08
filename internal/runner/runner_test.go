package runner_test

import (
	"strings"
	"testing"

	"github.com/mickamy/gotcha/internal/runner"
)

func TestParseEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantTotal   int
		wantPassed  int
		wantFailed  int
		wantSkipped int
		wantOK      bool
		wantFails   []string
	}{
		{
			name: "all pass",
			input: lines(
				`{"Action":"run","Package":"pkg","Test":"TestA"}`,
				`{"Action":"output","Package":"pkg","Test":"TestA","Output":"ok\n"}`,
				`{"Action":"pass","Package":"pkg","Test":"TestA"}`,
				`{"Action":"run","Package":"pkg","Test":"TestB"}`,
				`{"Action":"pass","Package":"pkg","Test":"TestB"}`,
			),
			wantTotal:  2,
			wantPassed: 2,
			wantOK:     true,
		},
		{
			name: "one failure",
			input: lines(
				`{"Action":"run","Package":"pkg","Test":"TestA"}`,
				`{"Action":"output","Package":"pkg","Test":"TestA","Output":"FAIL\n"}`,
				`{"Action":"fail","Package":"pkg","Test":"TestA"}`,
				`{"Action":"run","Package":"pkg","Test":"TestB"}`,
				`{"Action":"pass","Package":"pkg","Test":"TestB"}`,
			),
			wantTotal:  2,
			wantPassed: 1,
			wantFailed: 1,
			wantOK:     false,
			wantFails:  []string{"pkg/TestA"},
		},
		{
			name: "skip",
			input: lines(
				`{"Action":"run","Package":"pkg","Test":"TestA"}`,
				`{"Action":"skip","Package":"pkg","Test":"TestA"}`,
			),
			wantTotal:   1,
			wantSkipped: 1,
			wantOK:      true,
		},
		{
			name: "package-level events ignored",
			input: lines(
				`{"Action":"pass","Package":"pkg"}`,
				`{"Action":"fail","Package":"pkg2"}`,
			),
			wantTotal: 0,
			wantOK:    true,
		},
		{
			name:       "empty input",
			input:      "",
			wantTotal:  0,
			wantOK:     true,
		},
		{
			name: "invalid json lines skipped",
			input: lines(
				`not json`,
				`{"Action":"pass","Package":"pkg","Test":"TestA"}`,
			),
			wantTotal:  1,
			wantPassed: 1,
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := runner.ParseEvents(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Total != tt.wantTotal {
				t.Errorf("Total: got %d, want %d", result.Total, tt.wantTotal)
			}
			if result.Passed != tt.wantPassed {
				t.Errorf("Passed: got %d, want %d", result.Passed, tt.wantPassed)
			}
			if result.Failed != tt.wantFailed {
				t.Errorf("Failed: got %d, want %d", result.Failed, tt.wantFailed)
			}
			if result.Skipped != tt.wantSkipped {
				t.Errorf("Skipped: got %d, want %d", result.Skipped, tt.wantSkipped)
			}
			if result.OK != tt.wantOK {
				t.Errorf("OK: got %v, want %v", result.OK, tt.wantOK)
			}

			if len(result.FailedTests) != len(tt.wantFails) {
				t.Fatalf("FailedTests: got %d, want %d", len(result.FailedTests), len(tt.wantFails))
			}
			for i, id := range result.FailedTests {
				if got := id.String(); got != tt.wantFails[i] {
					t.Errorf("FailedTests[%d]: got %q, want %q", i, got, tt.wantFails[i])
				}
			}
		})
	}
}

func TestParseEvents_CapturesOutput(t *testing.T) {
	t.Parallel()

	input := lines(
		`{"Action":"run","Package":"pkg","Test":"TestA"}`,
		`{"Action":"output","Package":"pkg","Test":"TestA","Output":"line 1\n"}`,
		`{"Action":"output","Package":"pkg","Test":"TestA","Output":"line 2\n"}`,
		`{"Action":"fail","Package":"pkg","Test":"TestA"}`,
	)

	result, err := runner.ParseEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id := runner.TestID{Package: "pkg", Test: "TestA"}
	out, ok := result.Output[id]
	if !ok {
		t.Fatal("expected output for TestA")
	}
	if len(out) != 2 {
		t.Fatalf("output lines: got %d, want 2", len(out))
	}
	if out[0] != "line 1\n" {
		t.Errorf("output[0]: got %q, want %q", out[0], "line 1\n")
	}
	if out[1] != "line 2\n" {
		t.Errorf("output[1]: got %q, want %q", out[1], "line 2\n")
	}
}

func TestTestID_String(t *testing.T) {
	t.Parallel()

	id := runner.TestID{Package: "github.com/foo/bar", Test: "TestSomething"}
	if got := id.String(); got != "github.com/foo/bar/TestSomething" {
		t.Errorf("got %q, want %q", got, "github.com/foo/bar/TestSomething")
	}
}

func lines(ss ...string) string {
	return strings.Join(ss, "\n") + "\n"
}
