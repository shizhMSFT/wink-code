// Package llm provides retry logic with exponential backoff
package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/shizhMSFT/wink-code/internal/logging"
)

const (
	maxRetries     = 3
	initialBackoff = 1 * time.Second
	maxBackoff     = 10 * time.Second
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     maxRetries,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
	}
}

// WithRetry executes a function with exponential backoff retry
func WithRetry(ctx context.Context, config *RetryConfig, fn func() error) error {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Last attempt - don't retry
		if attempt == config.MaxRetries {
			break
		}

		// Log retry attempt
		logging.Warn("Request failed, retrying",
			"attempt", attempt+1,
			"max_retries", config.MaxRetries,
			"backoff_ms", backoff.Milliseconds(),
			"error", err,
		)

		// Wait with backoff
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
			// Calculate next backoff (exponential)
			backoff *= 2
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
