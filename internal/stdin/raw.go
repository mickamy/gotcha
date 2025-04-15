package stdin

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	fd       = int(os.Stdin.Fd())
	oldState *term.State
	raw      bool
)

func EnterRawMode() error {
	var err error
	oldState, err = term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to enter raw mode: %w", err)
	}
	raw = true
	return nil
}

func ExitRawMode() error {
	if raw {
		raw = false
		if err := term.Restore(fd, oldState); err != nil {
			return fmt.Errorf("failed to restore terminal: %w", err)
		}
	}
	return nil
}
