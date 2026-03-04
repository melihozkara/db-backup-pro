//go:build !windows

package main

// attachConsole is a no-op on non-Windows platforms.
// macOS and Linux already have a console when launched from terminal.
func attachConsole() {}
