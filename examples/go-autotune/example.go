// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Example CLI which can show GOMAXPROCS and GOMEMLIMIT values.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"syscall"
	"time"

	_ "embed"

	_ "github.com/tprasadtp/go-autotune"
)

//go:embed favicon.ico
var favicon []byte

func info(w io.Writer) {
	fmt.Fprintf(w, "GOOS       : %s\n", runtime.GOOS)
	fmt.Fprintf(w, "GOARCH     : %s\n", runtime.GOARCH)
	fmt.Fprintf(w, "GOMAXPROCS : %d\n", runtime.GOMAXPROCS(-1))
	fmt.Fprintf(w, "NumCPU     : %d\n", runtime.NumCPU())
	fmt.Fprintf(w, "GOMEMLIMIT : %d\n", debug.SetMemoryLimit(-1))
}

func main() {
	var addr string
	var wg sync.WaitGroup

	// Parse flags.
	flag.StringVar(&addr, "listen", "", "listen address")
	flag.Parse()

	// If server is not specified, but PORT is set, listen on all interfaces
	// on that port.
	if addr == "" {
		if v := os.Getenv("PORT"); v != "" {
			_, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				slog.Error("Invalid PORT",
					slog.String("PORT", v), slog.Any("err", err))
				os.Exit(1)
			}
			addr = fmt.Sprintf(":%v", v)
		}
	}

	// Server is not specified, just write to stdout.
	if addr == "" {
		info(os.Stdout)
		return
	}

	// Start server if an address is specified.
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Simple HTTP server which returns current values for GOMAXPROCS and GOMEMLIMIT.
	// Its output format is not subject to semver compatibility guarantees.
	srv := http.Server{
		Addr:              addr,
		ReadHeaderTimeout: time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				slog.Warn("Request",
					slog.String("client", r.RemoteAddr),
					slog.String("method", r.Method),
					slog.Any("url", r.URL),
					slog.Int("status", http.StatusMethodNotAllowed),
				)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			switch r.URL.Path {
			case "/", "":
				slog.Info("Request",
					slog.String("client", r.RemoteAddr),
					slog.String("method", r.Method),
					slog.Any("url", r.URL),
					slog.Int("status", http.StatusOK),
				)
				w.Header().Add("Content-Type", "text/plain")
				info(w)
			case "/favicon.ico":
				slog.Info("Request",
					slog.String("client", r.RemoteAddr),
					slog.String("method", r.Method),
					slog.Any("url", r.URL),
					slog.Int("status", http.StatusOK),
				)
				w.Header().Add("Content-Type", "image/x-icon")
				_, _ = w.Write(favicon)
			default:
				slog.Warn("Request",
					slog.String("client", r.RemoteAddr),
					slog.String("method", r.Method),
					slog.Any("url", r.URL),
					slog.Int("status", http.StatusNotFound),
				)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	}

	// Starts a go routine which handles server shutdown.
	wg.Add(1)
	go func() {
		var err error
		defer wg.Done()
		for {
			select {
			// on cancel, stop the server and return.
			case <-ctx.Done():
				slog.Info("Stopping server", "server", srv.Addr)
				err = srv.Shutdown(context.Background())
				if err != nil {
					slog.Error("Failed to shutdown server", slog.Any("err", err))
				}
				return
			}
		}
	}()

	// Start server if context is not already cancelled.
	if ctx.Err() == nil {
		slog.Info("Starting server", slog.String("server", srv.Addr))
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start the server", slog.Any("err", err))
			wg.Wait() // allow shutdown to complete.
			os.Exit(1)
		}
	} else {
		slog.Warn("Not starting server, context already cancelled", slog.String("server", srv.Addr))
	}

	wg.Wait()
	slog.Info("Server stopped", slog.String("server", srv.Addr))
}
