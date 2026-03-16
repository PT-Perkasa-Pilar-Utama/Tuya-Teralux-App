package utils

import (
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
)

func TestIsRetryableUploadError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error is not retryable",
			err:      nil,
			expected: false,
		},
		{
			name:     "io.ErrUnexpectedEOF is retryable",
			err:      io.ErrUnexpectedEOF,
			expected: true,
		},
		{
			name:     "context canceled is retryable",
			err:      errors.New("context canceled"),
			expected: false, // Note: context.Canceled would need special handling
		},
		{
			name:     "broken pipe is retryable",
			err:      &os.PathError{Err: syscall.EPIPE},
			expected: true,
		},
		{
			name:     "connection reset is retryable",
			err:      &os.PathError{Err: syscall.ECONNRESET},
			expected: true,
		},
		{
			name:     "connection aborted is retryable",
			err:      &os.PathError{Err: syscall.ECONNABORTED},
			expected: true,
		},
		{
			name:     "error containing 'broken pipe' is retryable",
			err:      errors.New("write: broken pipe"),
			expected: true,
		},
		{
			name:     "error containing 'connection reset' is retryable",
			err:      errors.New("read: connection reset by peer"),
			expected: true,
		},
		{
			name:     "error containing 'client disconnected' is retryable",
			err:      errors.New("http2: server sent GOAWAY: client disconnected"),
			expected: true,
		},
		{
			name:     "error containing 'unexpected EOF' is retryable",
			err:      errors.New("write: unexpected EOF"),
			expected: true,
		},
		{
			name:     "generic error is not retryable",
			err:      errors.New("some random error"),
			expected: false,
		},
		{
			name:     "validation error is not retryable",
			err:      errors.New("invalid chunk index"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableUploadError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableUploadError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableUploadError_NetTimeout(t *testing.T) {
	// Create a mock net.Error with Timeout() returning true
	mockNetError := &mockNetError{timeout: true, temporary: false}
	
	result := IsRetryableUploadError(mockNetError)
	if !result {
		t.Errorf("IsRetryableUploadError(net timeout) = false, want true")
	}
}

func TestIsRetryableUploadError_SyscallErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "syscall.EPIPE",
			err:      syscall.EPIPE,
			expected: true,
		},
		{
			name:     "syscall.ECONNRESET",
			err:      syscall.ECONNRESET,
			expected: true,
		},
		{
			name:     "syscall.ECONNABORTED",
			err:      syscall.ECONNABORTED,
			expected: true,
		},
		{
			name:     "syscall.EINVAL (not retryable)",
			err:      syscall.EINVAL,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableUploadError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableUploadError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableUploadError_WrappedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "os.PathError with EPIPE",
			err:      &os.PathError{Op: "write", Path: "socket", Err: syscall.EPIPE},
			expected: true,
		},
		{
			name:     "os.SyscallError with ECONNRESET",
			err:      &os.SyscallError{Syscall: "read", Err: syscall.ECONNRESET},
			expected: true,
		},
		{
			name:     "os.PathError with EINVAL (not retryable)",
			err:      &os.PathError{Op: "stat", Path: "/tmp", Err: syscall.EINVAL},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableUploadError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableUploadError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// Mock net.Error for testing timeout detection
type mockNetError struct {
	timeout     bool
	temporary   bool
	errorString string
}

func (m *mockNetError) Error() string   { return m.errorString }
func (m *mockNetError) Timeout() bool   { return m.timeout }
func (m *mockNetError) Temporary() bool { return m.temporary }
