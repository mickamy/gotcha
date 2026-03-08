package watcher_test

import (
	"testing"
	"time"

	"github.com/mickamy/gotcha/internal/watcher"
)

func TestDebounceLoop_SingleSignal(t *testing.T) {
	t.Parallel()

	signal := make(chan struct{}, 1)
	trigger := make(chan struct{}, 1)
	stop := make(chan struct{})

	go watcher.DebounceLoop(signal, stop, trigger)
	defer close(stop)

	signal <- struct{}{}

	select {
	case <-trigger:
		// ok
	case <-time.After(watcher.DebounceDelay * 3):
		t.Fatal("trigger not received after single signal")
	}
}

func TestDebounceLoop_ResetsOnSubsequentSignals(t *testing.T) {
	t.Parallel()

	signal := make(chan struct{}, 8)
	trigger := make(chan struct{}, 1)
	stop := make(chan struct{})

	go watcher.DebounceLoop(signal, stop, trigger)
	defer close(stop)

	// Send first signal.
	signal <- struct{}{}

	// Send additional signals during the debounce window to reset the timer.
	watcher.Sleep(0.3)
	signal <- struct{}{}
	watcher.Sleep(0.3)
	signal <- struct{}{}

	// At this point ~0.6x has passed since the last signal.
	// The trigger should NOT have fired yet (timer was reset).
	select {
	case <-trigger:
		t.Fatal("trigger fired too early; debounce should have reset")
	default:
	}

	// Wait for the full debounce delay after the last signal.
	select {
	case <-trigger:
		// ok
	case <-time.After(watcher.DebounceDelay * 3):
		t.Fatal("trigger not received after debounce settled")
	}
}

func TestDebounceLoop_MultipleBursts(t *testing.T) {
	t.Parallel()

	signal := make(chan struct{}, 8)
	trigger := make(chan struct{}, 1)
	stop := make(chan struct{})

	go watcher.DebounceLoop(signal, stop, trigger)
	defer close(stop)

	for i := 0; i < 3; i++ {
		signal <- struct{}{}

		select {
		case <-trigger:
			// ok
		case <-time.After(watcher.DebounceDelay * 3):
			t.Fatalf("burst %d: trigger not received", i)
		}

		// Small gap between bursts.
		watcher.Sleep(0.2)
	}
}

func TestDebounceLoop_StopDuringWait(t *testing.T) {
	t.Parallel()

	signal := make(chan struct{}, 1)
	trigger := make(chan struct{}, 1)
	stop := make(chan struct{})

	done := make(chan struct{})
	go func() {
		watcher.DebounceLoop(signal, stop, trigger)
		close(done)
	}()

	signal <- struct{}{}
	watcher.Sleep(0.3)
	close(stop)

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("debounceLoop did not exit after stop")
	}

	// Trigger should NOT have fired.
	select {
	case <-trigger:
		t.Fatal("trigger should not fire after stop")
	default:
	}
}

func TestDebounceLoop_StopBeforeSignal(t *testing.T) {
	t.Parallel()

	signal := make(chan struct{}, 1)
	trigger := make(chan struct{}, 1)
	stop := make(chan struct{})

	done := make(chan struct{})
	go func() {
		watcher.DebounceLoop(signal, stop, trigger)
		close(done)
	}()

	close(stop)

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("debounceLoop did not exit on stop")
	}
}

func TestDebounceLoop_TriggerBackpressure(t *testing.T) {
	t.Parallel()

	signal := make(chan struct{}, 1)
	trigger := make(chan struct{}) // unbuffered, no reader
	stop := make(chan struct{})

	go watcher.DebounceLoop(signal, stop, trigger)
	defer close(stop)

	signal <- struct{}{}

	// debounceLoop should not block forever when trigger is full;
	// it uses a non-blocking send.
	select {
	case <-trigger:
		// consumed
	case <-time.After(watcher.DebounceDelay * 3):
		t.Fatal("debounceLoop blocked on full trigger channel")
	}
}

func TestRunLoop_ExecutesAndStops(t *testing.T) {
	t.Parallel()

	trigger := make(chan struct{}, 1)
	stop := make(chan struct{})
	count := 0

	done := make(chan struct{})
	go func() {
		watcher.RunLoop(trigger, stop, func() { count++ })
		close(done)
	}()

	trigger <- struct{}{}
	trigger <- struct{}{}
	watcher.Sleep(0.2)
	close(stop)

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("runLoop did not exit after stop")
	}

	if count != 2 {
		t.Errorf("count: got %d, want 2", count)
	}
}
