package utils

import (
	"io"
	"net"
	"os"
	"strings"
	"syscall"
)

// isRetryableUploadError determines if an error during chunk upload is retryable.
// Retryable errors include:
// - io.ErrUnexpectedEOF (client/proxy dropped connection)
// - context.Canceled (request canceled)
// - timeout errors (net.Error with Timeout())
// - broken pipe / connection reset / client disconnected
func IsRetryableUploadError(err error) bool {
	if err == nil {
		return false
	}

	// Direct io.ErrUnexpectedEOF - client/proxy dropped connection mid-body
	if err == io.ErrUnexpectedEOF {
		return true
	}

	// Check for timeout errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return true
		}
	}

	// Check error message for common retryable patterns
	errMsg := err.Error()
	retryablePatterns := []string{
		"broken pipe",
		"connection reset",
		"client disconnected",
		"connection closed",
		"unexpected eof",
		"use of closed network connection",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			return true
		}
	}

	// Check for syscall errors (EPIPE, ECONNRESET, ECONNABORTED)
	if isSyscallError(err) {
		return true
	}

	return false
}

// isSyscallError checks if the error is a low-level syscall error indicating connection issues
func isSyscallError(err error) bool {
	for {
		switch e := err.(type) {
		case *os.PathError:
			err = e.Err
		case *os.SyscallError:
			err = e.Err
		case syscall.Errno:
			return e == syscall.EPIPE || e == syscall.ECONNRESET || e == syscall.ECONNABORTED
		default:
			return false
		}
	}
}
