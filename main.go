package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime/debug"
	"os"
	"os/signal"
	"syscall"

	"pkb/internal/kb"
	"pkb/internal/server"
	"pkb/internal/watcher"
)

// Set via -ldflags at build time; defaults for plain "go build".
var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	var rev string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev != "" {
		if len(rev) > 8 {
			rev = rev[:8]
		}
		version = "dev (" + rev
		if dirty {
			version += "-dirty"
		}
		version += ")"
	}
}

func main() {
	var (
		dir     = flag.String("C", ".", "knowledge base directory (default: current directory)")
		addr    = flag.String("addr", "127.0.0.1:4242", "listen address")
		showVer = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Println("pkb", version)
		return
	}

	kbase, err := kb.New(*dir)
	if err != nil {
		log.Fatalf("pkb: %v", err)
	}

	srv, err := server.New(kbase, *addr)
	if err != nil {
		log.Fatalf("pkb: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start the file watcher.
	w := watcher.New(srv.Hub(), kbase)
	go w.Start(ctx)

	// Start the HTTP server (blocks until ctx is cancelled).
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("pkb: %v", err)
	}
}
