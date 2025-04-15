package paths

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func ListPackages(in []string) ([]string, error) {
	cmd := exec.Command("go", append([]string{"list"}, in...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func ModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get module path: %w", err)
	}

	return strings.TrimSpace(out.String()), nil
}
