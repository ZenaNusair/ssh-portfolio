package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

const host = "0.0.0.0"

func main() {
	// Port is configurable via env so the container can listen on 22
	// while local dev uses 2222 (avoids needing sudo on macOS).
	port := os.Getenv("PORT")
	if port == "" {
		port = "2222"
	}

	hostKeyPath := os.Getenv("HOST_KEY_PATH")
	if hostKeyPath == "" {
		hostKeyPath = ".ssh/host_key"
	}

	// If HOST_KEY_B64 is set (production via Fly secrets), decode it and
	// write it to hostKeyPath so the SSH fingerprint stays stable across
	// deploys. Falls back to whatever's already at hostKeyPath otherwise.
	if b64 := os.Getenv("HOST_KEY_B64"); b64 != "" {
		key, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			log.Fatalf("invalid HOST_KEY_B64: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(hostKeyPath), 0o700); err != nil {
			log.Fatalf("mkdir host key dir: %v", err)
		}
		if err := os.WriteFile(hostKeyPath, key, 0o600); err != nil {
			log.Fatalf("write host key: %v", err)
		}
	}

	s, err := wish.NewServer(
		wish.WithAddress(host+":"+port),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("could not create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting SSH server on %s:%s", host, port)

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	log.Println("server stopped")
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	w := pty.Window.Width
	h := pty.Window.Height
	if w == 0 {
		w = 120
	}
	if h == 0 {
		h = 40
	}

	// Force truecolor for this SSH session so neon colors actually render
	// (without this, lipgloss detects the server's stdout and renders monochrome).
	renderer := bubbletea.MakeRenderer(s)
	renderer.SetColorProfile(termenv.TrueColor)
	lipgloss.SetDefaultRenderer(renderer)

	m := NewModel(w, h)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
