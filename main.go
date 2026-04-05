package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pkb/internal/kb"
	"pkb/internal/server"
	"pkb/internal/watcher"
)

func main() {
	var (
		dir  = flag.String("C", ".", "knowledge base directory (default: current directory)")
		addr = flag.String("addr", "127.0.0.1:4242", "listen address")
	)
	flag.Parse()

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
