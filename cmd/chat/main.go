package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go-chat/internal/app"
	"go-chat/internal/tui"
	"go-chat/internal/tunnel"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	versionFlag := flag.Bool("version", false, "print version")
	serveFlag := flag.Bool("serve", false, "run as headless relay (no TUI)")
	tunnelAddr := flag.String("tunnel", "", "run tunnel server on this address (e.g. :1234)")
	nameFlag := flag.String("name", "", "display name (will prompt if not set)")
	flag.Parse()

	if *versionFlag {
		fmt.Println("go-chat v0.1.2")
		return
	}

	if *serveFlag {
		runRelay(*configPath, *tunnelAddr)
		return
	}

	a, err := app.New(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *nameFlag != "" {
		name := strings.TrimSpace(*nameFlag)
		if name == "me" || strings.HasPrefix(name, "me_") {
			fmt.Fprintf(os.Stderr, "Error: '%s' is a reserved name, choose another\n", name)
			os.Exit(1)
		}
		a.SetDisplayName(name)
	}

	p := tea.NewProgram(
		tui.NewModel(a),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		a.Close()
		os.Exit(1)
	}

	if err := a.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}
}

func runRelay(cfgPath, tunnelAddr string) {
	fmt.Println("go-chat relay mode")

	a, err := app.New(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if tunnelAddr != "" {
		fmt.Printf("Starting tunnel server on %s\n", tunnelAddr)
		go func() {
			if err := tunnel.RunServer(tunnelAddr); err != nil {
				fmt.Fprintf(os.Stderr, "Tunnel server error: %v\n", err)
			}
		}()
	}

	peerID := a.PeerID()
	fmt.Printf("Relay Peer ID: %s\n", peerID)
	fmt.Println("Share this address with peers:")
	for _, addr := range a.AllAddrs() {
		fmt.Printf("  %s\n", addr)
	}
	fmt.Println()
	fmt.Println("Peers connect with: /connect <address>")
	if tunnelAddr != "" {
		fmt.Printf("Tunnel server listening on %s\n", tunnelAddr)
		fmt.Println("Chat clients use: /tunnel <server-ip>:<port>")
	}
	fmt.Println("Press Ctrl+C to stop.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		fmt.Println("\nShutting down...")
	case <-ctx.Done():
	}

	a.Close()
}
