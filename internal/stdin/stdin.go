package stdin

import (
	"os"
	"slices"
)

type KeyPressDownEvent struct {
	Key string // non-empty if a valid key from `keys` slice
	EOF bool   // true if stdin closed (e.g. ctrl+d)
}

func Listen(keys []string, trigger chan<- KeyPressDownEvent) error {
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
		}

		key := string(buf[0])
		if slices.Contains(keys, key) {
			trigger <- KeyPressDownEvent{Key: key}
		}
	}
}
