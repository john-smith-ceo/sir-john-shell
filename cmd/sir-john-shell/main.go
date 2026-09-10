package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sir-john-shell/internal/acp"
	"sir-john-shell/internal/server"
)

func main() {
	var (
		addr  = flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
		cwd   = flag.String("cwd", ".", "Working directory for the session")
		devin = flag.String("devin", "devin", "Path to devin CLI binary")
	)
	flag.Parse()

	absCwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}
	if *cwd != "." {
		absCwd = *cwd
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := acp.NewClient(ctx, *devin, nil)
	if err != nil {
		log.Fatalf("acp client: %v", err)
	}
	defer client.Close()

	log.Printf("ACP client started")

	initRes, err := client.Initialize(ctx)
	if err != nil {
		log.Fatalf("acp initialize: %v", err)
	}
	log.Printf("ACP initialized: %s v%s", initRes.AgentInfo.Title, initRes.AgentInfo.Version)

	session, err := client.NewSession(ctx, absCwd)
	if err != nil {
		log.Fatalf("acp session/new: %v", err)
	}
	log.Printf("ACP session: %s", session.SessionID)

	srv := server.New(*addr, absCwd, client, session.SessionID)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down")
		cancel()
		_ = client.Close()
	}()

	if err := srv.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
