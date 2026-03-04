package main

import (
	"dbbackup/internal/config"
	"dbbackup/internal/server"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check for "serve" subcommand
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runWebMode(os.Args[2:])
		return
	}

	// Proactive WebView2 check (Windows only — macOS/Linux always true)
	if !isWebView2Available() {
		fmt.Println("WebView2 Runtime bulunamadı.")
		fmt.Println("Web moduna geçiliyor...")
		runWebMode(nil)
		return
	}

	// WebView2 mevcut → masaüstü modunu dene
	runDesktopMode()
}

// runDesktopMode starts the Wails desktop application
func runDesktopMode() {
	app := NewApp()

	appOpts := &options.App{
		Title:     "DB Backup Pro",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	}

	// Windows-specific: WebView2 crash onleme
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = os.TempDir()
		}
		appOpts.Windows = &windows.Options{
			// Antivirus DLL injection crash'ini onle
			WebviewDisableRendererCodeIntegrity: true,
			// Yetki sorunlarini onle (admin/normal kullanici)
			WebviewUserDataPath: filepath.Join(appData, "DBBackupPro"),
		}
	}

	err := wails.Run(appOpts)

	if err != nil {
		// WebView2 not available or other Wails error → fallback to web mode
		fmt.Printf("Masaüstü modu başlatılamadı: %s\n", err.Error())
		fmt.Println("Web moduna geçiliyor...")
		runWebMode(nil)
	}
}

// runWebMode starts the HTTP server for browser-based access
func runWebMode(args []string) {
	// On Windows, allocate a console window so fmt.Println output is visible
	// (Wails builds as GUI app with no console)
	attachConsole()

	// Parse CLI flags
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portFlag := fs.Int("port", 0, "HTTP server port")
	hostFlag := fs.String("host", "", "HTTP server host")
	if args != nil {
		fs.Parse(args)
	}

	// Load config from file
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Config yükleme hatası: %s\n", err.Error())
		os.Exit(1)
	}

	// CLI flags override config file
	if *portFlag != 0 {
		cfg.Server.Port = *portFlag
	}
	if *hostFlag != "" {
		cfg.Server.Host = *hostFlag
	}

	// Initialize app and services
	app := NewApp()

	// Create SSE hub and wire it as the event emitter
	sseHub := server.NewSSEHub()
	app.SetEventEmitter(sseHub.Broadcast)

	// Initialize backend services (DB, scheduler, backup, etc.)
	if err := app.initServices(); err != nil {
		fmt.Printf("Servisler başlatılamadı: %s\n", err.Error())
		os.Exit(1)
	}
	defer app.Shutdown()

	// Create and start HTTP server
	srv := server.New(app, assets, sseHub)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := srv.Start(addr); err != nil {
		fmt.Printf("Sunucu hatası: %s\n", err.Error())
		os.Exit(1)
	}
}
