package watcher

import "time"

// DebounceDelay is exported for testing.
const DebounceDelay = debounceDelay

// DebounceLoop is exported for testing.
func DebounceLoop(signal <-chan struct{}, stop <-chan struct{}, trigger chan<- struct{}) {
	debounceLoop(signal, stop, trigger)
}

// RunLoop is exported for testing.
func RunLoop(trigger <-chan struct{}, stop <-chan struct{}, fn func()) {
	runLoop(trigger, stop, fn)
}

// Sleep is a test helper that sleeps for a fraction of the debounce delay.
func Sleep(fraction float64) {
	time.Sleep(time.Duration(float64(debounceDelay) * fraction))
}
