package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"sir-john-shell/internal/acp"
	"sir-john-shell/internal/cli"
	"sir-john-shell/internal/config"
	"sir-john-shell/internal/server"
)

func main() {
	var (
		addr    = flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
		cwd     = flag.String("cwd", ".", "Working directory for the session")
		devin   = flag.String("devin", "devin", "Path to devin CLI binary")
		mode    = flag.String("mode", "cli", "Run mode: cli, web, or headless")
		logMode = flag.String("log", "off", "Theme-flow diagnostics: on or off")
		devMode = flag.Bool("dev", false, "Enable UX/UI polygon controls")
	)
	flag.Parse()
	logDir := ""
	switch *logMode {
	case "on":
		logDir = "logs"
	case "off", "":
	default:
		log.Fatalf("invalid -log value %q: use on or off", *logMode)
	}

	absCwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}
	if *cwd != "." {
		absCwd = *cwd
	}

	cfg, err := config.Load()
	if err != nil {
		log.Printf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := acp.NewClient(ctx, *devin, nil)
	if err != nil {
		log.Fatalf("acp client: %v", err)
	}
	defer client.Close()

	log.Printf("ACP client started")

	initCtx, initCancel := context.WithTimeout(ctx, 15*time.Second)
	initRes, err := client.Initialize(initCtx)
	initCancel()
	if err != nil {
		log.Fatalf("acp initialize: %v", err)
	}
	log.Printf("ACP initialized: %s v%s", initRes.AgentInfo.Title, initRes.AgentInfo.Version)

	sessionCtx, sessionCancel := context.WithTimeout(ctx, 30*time.Second)
	session, err := client.NewSession(sessionCtx, absCwd)
	sessionCancel()
	if err != nil {
		log.Fatalf("acp session/new: %v", err)
	}
	log.Printf("ACP session: %s", session.SessionID)

	srv := server.New(*addr, absCwd, cfg.User, client, session.SessionID, logDir, cancel)

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

	select {
	case <-srv.Ready():
	case err := <-serverErrCh:
		if err != nil {
			log.Fatalf("server: %v", err)
		}
		log.Fatalf("server stopped before it became ready")
	case <-time.After(5 * time.Second):
		log.Fatalf("server did not start in time")
	}

	switch *mode {
	case "web":
		log.Printf("Open http://%s", *addr)
		url := fmt.Sprintf("http://%s", *addr)
		if *devMode {
			url += "?dev=1"
		}
		_ = openBrowser(url)
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

func openBrowser(url string) error {
	for _, cmd := range []string{"xdg-open", "open"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return exec.Command(cmd, url).Start()
		}
	}
	return fmt.Errorf("no browser opener found")
}
