// Package ui provides user interface components
package ui

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"
)

// ProgressIndicator displays a spinner with elapsed time during long operations
type ProgressIndicator struct {
	writer    io.Writer
	message   string
	startTime time.Time
	stopChan  chan bool
	done      bool
	isTTY     bool
}

// Spinner frames for animation
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(message string) *ProgressIndicator {
	// Check if output is a TTY (terminal)
	isTTY := term.IsTerminal(int(os.Stderr.Fd()))

	return &ProgressIndicator{
		writer:   os.Stderr,
		message:  message,
		stopChan: make(chan bool),
		isTTY:    isTTY,
	}
}

// Start begins displaying the progress indicator
func (p *ProgressIndicator) Start() {
	if !p.isTTY {
		// In non-TTY environments, just print the message once
		fmt.Fprintf(p.writer, "%s...\n", p.message)
		return
	}

	p.startTime = time.Now()
	go p.spin()
}

// Stop stops the progress indicator
func (p *ProgressIndicator) Stop() {
	if p.done || !p.isTTY {
		return
	}

	p.done = true
	p.stopChan <- true

	// Clear the line
	fmt.Fprintf(p.writer, "\r\033[K")
}

// Update changes the message displayed by the progress indicator
func (p *ProgressIndicator) Update(message string) {
	if !p.isTTY {
		fmt.Fprintf(p.writer, "%s...\n", message)
		return
	}

	p.message = message
}

// spin runs the spinner animation
func (p *ProgressIndicator) spin() {
	frameIdx := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			elapsed := time.Since(p.startTime)
			frame := spinnerFrames[frameIdx%len(spinnerFrames)]
			frameIdx++

			// Format: "⠋ Message... (3.2s)"
			fmt.Fprintf(p.writer, "\r%s %s... (%s)", frame, p.message, formatDuration(elapsed))
		}
	}
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}
