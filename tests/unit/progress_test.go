// Package unit contains unit tests for internal packages
package unit

import (
	"testing"
	"time"

	"github.com/shizhMSFT/wink-code/internal/ui"
)

func TestProgressIndicatorCreation(t *testing.T) {
	p := ui.NewProgressIndicator("Testing")
	if p == nil {
		t.Fatal("Expected progress indicator to be created")
	}
}

func TestProgressIndicatorStartStopQuick(t *testing.T) {
	p := ui.NewProgressIndicator("Testing operation")

	// Start and immediately stop
	p.Start()
	p.Stop()

	// Should not panic or error
}

func TestProgressIndicatorMultipleStops(t *testing.T) {
	p := ui.NewProgressIndicator("Testing")

	p.Start()

	// Multiple stops should be safe
	p.Stop()
	p.Stop()
	p.Stop()
}

func TestProgressIndicatorWithoutStart(t *testing.T) {
	p := ui.NewProgressIndicator("Testing")

	// Calling Stop without Start should be safe
	p.Stop()
}

func TestProgressIndicatorUpdate(t *testing.T) {
	p := ui.NewProgressIndicator("Initial message")

	p.Start()
	p.Update("Updated message")
	p.Stop()

	// Should not panic
}

func TestProgressIndicatorShortDuration(t *testing.T) {
	// Test with a very short duration
	p := ui.NewProgressIndicator("Quick test")

	p.Start()
	time.Sleep(50 * time.Millisecond)
	p.Stop()
}

func TestProgressIndicatorNonTTY(t *testing.T) {
	// When not in a TTY, progress indicator should just print once
	// This is tested implicitly in CI environments

	p := ui.NewProgressIndicator("Non-TTY test")
	p.Start()
	p.Update("Updated")
	p.Stop()

	// Should complete without hanging
}

func TestProgressIndicatorRapidCycle(t *testing.T) {
	// Test rapid start/stop cycles
	for i := 0; i < 5; i++ {
		p := ui.NewProgressIndicator("Rapid test")
		p.Start()
		p.Stop()
	}
}
