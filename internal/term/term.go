package term

import (
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"golang.org/x/term"
)

// KeyEvent represents a keyboard input event.
type KeyEvent struct {
	Key string // non-empty if a recognized key was pressed
	EOF bool   // true on ctrl+d (stdin closed)
}

// RawMode manages terminal raw mode state.
type RawMode struct {
	fd    int
	state *term.State
	raw   bool
}

// NewRawMode creates a RawMode for the given file descriptor.
func NewRawMode(fd int) *RawMode {
	return &RawMode{fd: fd}
}

// Enter puts the terminal into raw mode.
func (r *RawMode) Enter() error {
	state, err := term.MakeRaw(r.fd)
	if err != nil {
		return fmt.Errorf("enter raw mode: %w", err)
	}
	r.state = state
	r.raw = true
	return nil
}

// Exit restores the terminal to its previous state.
func (r *RawMode) Exit() error {
	if !r.raw {
		return nil
	}
	r.raw = false
	if err := term.Restore(r.fd, r.state); err != nil {
		return fmt.Errorf("restore terminal: %w", err)
	}
	return nil
}

// Listen reads single bytes from stdin and sends KeyEvents for recognized keys.
// It returns when done is closed, ctrl+c/ctrl+d is received, or stdin reaches EOF.
//
// Note: the internal readLoop goroutine may outlive Listen because
// os.Stdin.Read is a blocking syscall with no cancellation mechanism.
// This is acceptable for a CLI tool where process exit reclaims all resources.
func Listen(keys []string, ch chan<- KeyEvent, done <-chan struct{}) {
	reads := make(chan byte, 1)

	go readLoop(reads)

	for {
		select {
		case <-done:
			return
		case b := <-reads:
			switch b {
			case 3: // ctrl+c
				ch <- KeyEvent{Key: "ctrl+c"}
				return
			case 4: // ctrl+d
				ch <- KeyEvent{EOF: true}
				return
			}

			key := string(b)
			if slices.Contains(keys, key) {
				ch <- KeyEvent{Key: key}
			}
		}
	}
}

func readLoop(out chan<- byte) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			out <- buf[0]
		}
		if err != nil {
			if err == io.EOF {
				out <- 4 // signal as ctrl+d
				return
			}
			// Avoid busy loop on persistent errors.
			time.Sleep(10 * time.Millisecond)
		}
	}
}
