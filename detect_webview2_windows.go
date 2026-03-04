//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// isWebView2Available checks if WebView2 Runtime is installed on Windows
func isWebView2Available() bool {
	// 1. Registry kontrolu (system-wide + per-user)
	regPaths := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BEF-335EB7735854}`},
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BEF-335EB7735854}`},
		{registry.CURRENT_USER, `Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BEF-335EB7735854}`},
	}

	for _, rp := range regPaths {
		k, err := registry.OpenKey(rp.root, rp.path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		ver, _, err := k.GetStringValue("pv")
		k.Close()

		if err == nil && ver != "" && ver != "0.0.0.0" {
			if !isVersionOK(ver) {
				fmt.Printf("WebView2 bulundu ama versiyon cok eski: v%s (min 118 gerekli)\n", ver)
				continue
			}
			fmt.Printf("WebView2 bulundu (registry): v%s\n", ver)
			return true
		}
	}

	// 2. Registry'de bulunamadiysa, dosya sistemi kontrolu yap
	// Bazi durumlarda WebView2 yuklu ama registry kaydi eksik olabiliyor
	runtimePaths := []string{
		`C:\Program Files (x86)\Microsoft\EdgeWebView\Application`,
		`C:\Program Files\Microsoft\EdgeWebView\Application`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Microsoft\EdgeWebView\Application`),
	}

	for _, basePath := range runtimePaths {
		if checkWebView2Files(basePath) {
			fmt.Printf("WebView2 bulundu (dosya sistemi): %s\n", basePath)
			return true
		}
	}

	return false
}

// checkWebView2Files checks if msedgewebview2.exe exists in the given base path
func checkWebView2Files(basePath string) bool {
	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return false
	}

	// Look for msedgewebview2.exe in subdirectories (version folders like 145.0.3800.82)
	matches, err := filepath.Glob(filepath.Join(basePath, "*", "msedgewebview2.exe"))
	if err != nil {
		return false
	}

	return len(matches) > 0
}

// isVersionOK checks if WebView2 version >= 118 (older versions crash on focus)
func isVersionOK(ver string) bool {
	parts := strings.Split(ver, ".")
	if len(parts) < 1 {
		return false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	return major >= 118
}
