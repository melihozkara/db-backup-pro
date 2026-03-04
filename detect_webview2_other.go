//go:build !windows && !linux

package main

// isWebView2Available on macOS always returns true (WebKit is built-in).
func isWebView2Available() bool {
	return true
}
