//go:build windows

package main

import (
	"os"
	"syscall"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole  = kernel32.NewProc("AllocConsole")
	procAttachConsole = kernel32.NewProc("AttachConsole")
)

// attachConsole attaches to the parent console (cmd/PowerShell) or creates a new one.
// Needed because Wails builds as a GUI app with no console window.
func attachConsole() {
	// Try to attach to parent process console (cmd.exe, PowerShell, etc.)
	r, _, _ := procAttachConsole.Call(^uintptr(0)) // ATTACH_PARENT_PROCESS = -1
	if r == 0 {
		// No parent console → create a new one (double-click scenario)
		procAllocConsole.Call()
	}

	// Redirect stdout/stderr to the (newly attached) console handles
	stdout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	stderr, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
	os.Stdout = os.NewFile(uintptr(stdout), "stdout")
	os.Stderr = os.NewFile(uintptr(stderr), "stderr")
}
