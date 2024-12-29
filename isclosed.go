package main

import (
	"errors"
	"net"
	"syscall"
)

// For checking errors when writing HTTP responses, we don't want to log i/o
// errors, but we do want to see other errors, e.g. about template execution.
func isClosed(err error) bool {
	return errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || isRemoteTLSError(err)
}

// A remote TLS client can send a message indicating failure, this makes it back to
// us as a write error.
func isRemoteTLSError(err error) bool {
	var netErr *net.OpError
	return errors.As(err, &netErr) && netErr.Op == "remote error"
}
