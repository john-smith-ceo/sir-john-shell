package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"sir-john-shell/internal/acp"
	"sir-john-shell/internal/cli"
	"sir-john-shell/internal/server"
)

func main() {
	var (
		addr  = flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
		cwd   = flag.String("cwd", ".", "Working directory for the session")
		devin = flag.String("devin", "devin", "Path to devin CLI binary")
		mode  = flag.String("mode", "cli", "Run mode: cli, web, or headless")
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

	serverErrCh := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	go func() {
		<-sigCh
		log.Println("shutting down")
		cancel()
		_ = client.Close()
	}()

	if !waitForServer(*addr, 5*time.Second) {
		log.Fatalf("server did not start in time")
	}

	switch *mode {
	case "web":
		log.Printf("Open http://%s", *addr)
		_ = openBrowser(fmt.Sprintf("http://%s", *addr))
		<-serverErrCh

	case "headless":
		log.Printf("Server running on http://%s", *addr)
		<-serverErrCh

	case "cli", "":
		if err := cli.Run(*addr, ctx); err != nil {
			log.Printf("cli: %v", err)
		}
		cancel()
		_ = client.Close()

	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

func waitForServer(addr string, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func openBrowser(url string) error {
	for _, cmd := range []string{"xdg-open", "open"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return exec.Command(cmd, url).Start()
		}
	}
	return fmt.Errorf("no browser opener found")
}
