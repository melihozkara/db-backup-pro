package server

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// Server holds the HTTP server and its dependencies
type Server struct {
	app        AppAPI
	httpServer *http.Server
	SSEHub     *SSEHub
	assets     fs.FS
}

// New creates a new HTTP server. The app must implement AppAPI (defined in handlers.go).
func New(app AppAPI, assets fs.FS, sseHub *SSEHub) *Server {
	return &Server{
		app:    app,
		SSEHub: sseHub,
		assets: assets,
	}
}

// Start starts the HTTP server and blocks until shutdown
func (s *Server) Start(addr string) error {
	mux := s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(apiTimeoutMiddleware(mux)),
		ReadTimeout:  0,  // SSE connections need unlimited read timeout
		WriteTimeout: 0,  // SSE connections need unlimited write timeout
		IdleTimeout:  120 * time.Second,
	}

	// Check if port is available by listening directly
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("Port %s kullanılıyor. --port ile farklı port belirtin", addr)
	}

	// Graceful shutdown on signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		fmt.Println("\nSunucu kapatılıyor...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}()

	url := fmt.Sprintf("http://%s", addr)
	fmt.Printf("\n  DB Backup Pro - Web Modu\n")
	fmt.Printf("  ─────────────────────────\n")
	fmt.Printf("  Adres: %s\n\n", url)

	// Open browser automatically
	go openBrowser(url)

	// Serve on the already-open listener (no port race)
	if err := s.httpServer.Serve(ln); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// openBrowser opens the given URL in the default browser
func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

// apiTimeoutMiddleware applies a timeout to API endpoints (excluding SSE)
func apiTimeoutMiddleware(next http.Handler) http.Handler {
	timeoutHandler := http.TimeoutHandler(next, 60*time.Second, `{"error":"request timeout"}`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SSE endpoint must not have a timeout
		if r.URL.Path == "/api/events" {
			next.ServeHTTP(w, r)
			return
		}
		timeoutHandler.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers with origin validation
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow requests from the same host (localhost/127.0.0.1 with any port)
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
