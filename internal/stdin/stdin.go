package stdin

import (
	"fmt"
	"os"
	"slices"

	"golang.org/x/term"
)

type KeyPressDownEvent struct {
	Key string
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
		_, err := os.Stdin.Read(buf)
		if err != nil {
			continue
		}
		key := string(buf[0])
		if !slices.Contains(keys, key) {
			continue
		}
		trigger <- KeyPressDownEvent{Key: key}
	}
}
