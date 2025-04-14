package stdin

import (
	"fmt"
	"os"
	"slices"

	"golang.org/x/term"
)

type KeyPressDownEvent struct {
	Key string // non-empty if a valid key from `keys` slice
	EOF bool   // true if stdin closed (e.g. ctrl+d)
}

func Listen(keys []string, trigger chan<- KeyPressDownEvent) error {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func(fd int, oldState *term.State) {
		_ = term.Restore(fd, oldState)
	}(fd, oldState)

	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		switch buf[0] {
		case 3: // ctrl+c
			trigger <- KeyPressDownEvent{Key: "ctrl+c"}
			return nil
		case 4: // ctrl+d
			trigger <- KeyPressDownEvent{EOF: true}
			return nil
		}

		key := string(buf[0])
		if slices.Contains(keys, key) {
			trigger <- KeyPressDownEvent{Key: key}
		}
	}
}
