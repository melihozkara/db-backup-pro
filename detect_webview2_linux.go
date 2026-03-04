//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// isWebView2Available on Linux checks if a display server and GTK/WebKit are available.
// Headless sunucularda (CentOS, AlmaLinux vb.) false donup web moduna gecer.
func isWebView2Available() bool {
	// 1. Display server var mi? (X11 veya Wayland)
	display := os.Getenv("DISPLAY")
	wayland := os.Getenv("WAYLAND_DISPLAY")
	if display == "" && wayland == "" {
		fmt.Println("Display server bulunamadi (headless ortam), web moduna geciliyor...")
		return false
	}

	// 2. WebKit2GTK kutuphanesi yuklu mu? (4.0 veya 4.1)
	err40 := exec.Command("pkg-config", "--exists", "webkit2gtk-4.0").Run()
	err41 := exec.Command("pkg-config", "--exists", "webkit2gtk-4.1").Run()
	if err40 != nil && err41 != nil {
		fmt.Println("webkit2gtk bulunamadi, web moduna geciliyor...")
		return false
	}

	fmt.Println("Linux masaustu ortami tespit edildi")
	return true
}
