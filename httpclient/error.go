package httpclient

import (
	"fmt"
	"net"
	"strings"
)

// Package errors
const (
	_ = iota
	ErrDefault
	ErrTimeout
	ErrRedirectPolicy
)

//Error Custom error
type Error struct {
	Code    int
	Message string
}

//Error Implement the error interface
func (this Error) Error() string {
	return fmt.Sprintf("httpclient #%d: %s", this.Code, this.Message)
}

func getErrorCode(err error) int {
	if err == nil {
		return 0
	}

	if e, ok := err.(*Error); ok {
		return e.Code
	}

	return ErrDefault
}

//IsTimeoutError Check a timeout error.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// TODO: does not work?
	if e, ok := err.(net.Error); ok && e.Timeout() {
		return true
	}

	// TODO: make it reliable
	if strings.Contains(err.Error(), "timeout") {
		return true
	}

	return false
}

//IsRedirectError Check a redirect error
func IsRedirectError(err error) bool {
	if err == nil {
		return false
	}

	// TODO: does not work?
	if getErrorCode(err) == ErrRedirectPolicy {
		return true
	}

	// TODO: make it reliable
	if strings.Contains(err.Error(), "redirect") {
		return true
	}

	return false
}
