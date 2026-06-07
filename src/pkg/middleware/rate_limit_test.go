package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedWindowLimiter(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	limiter := newFixedWindowLimiter(2, time.Minute)

	// First two requests in the window are allowed.
	ok, _ := limiter.allow("a", base)
	assert.True(t, ok)
	ok, _ = limiter.allow("a", base.Add(time.Second))
	assert.True(t, ok)

	// Third is denied with a positive retry-after.
	ok, retry := limiter.allow("a", base.Add(2*time.Second))
	assert.False(t, ok)
	assert.Greater(t, retry, time.Duration(0))

	// A different client is independent.
	ok, _ = limiter.allow("b", base.Add(2*time.Second))
	assert.True(t, ok)

	// After the window resets, the original client is allowed again.
	ok, _ = limiter.allow("a", base.Add(time.Minute+time.Second))
	assert.True(t, ok)
}
