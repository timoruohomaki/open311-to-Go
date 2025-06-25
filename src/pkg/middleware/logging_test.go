package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	logs []string
}

func (m *mockLogger) Debug(args ...interface{})                 {}
func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Info(args ...interface{})                  {}
func (m *mockLogger) Warn(args ...interface{})                  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{})                 {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}
func (m *mockLogger) Fatal(args ...interface{})                 {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) {}
func (m *mockLogger) Close() error                              { return nil }
func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}

func TestLoggingMiddleware_ApacheFormat(t *testing.T) {
	mockLog := &mockLogger{}
	mw := LoggingMiddleware(mockLog)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest("POST", "/api/v1/test", strings.NewReader(""))
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("Referer", "http://example.com/")
	rec := httptest.NewRecorder()

	mw(h).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// There should be one log entry
	assert.Len(t, mockLog.logs, 1)
	logLine := mockLog.logs[0]

	// Apache Combined Log Format regex (simplified)
	apacheRegex := regexp.MustCompile(`^.+ - - \[.+\] ".+ .+ .+" \d+ \d+ ".*" ".*"$`)
	assert.Regexp(t, apacheRegex, logLine)

	// Check that status and size are correct
	assert.Contains(t, logLine, "201")                 // Status
	assert.Contains(t, logLine, "TestAgent/1.0")       // User-Agent
	assert.Contains(t, logLine, "http://example.com/") // Referer
}
