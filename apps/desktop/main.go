// Command desktop is the Drift desktop backend: the Go core behind one
// loopback HTTP server (the SvelteKit UI from frontend/, the stream proxy and the
// RPC WebSocket). The Electron shell in electron/ starts it and opens a window on it.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 2 * time.Second

// announcePrefix starts the stdout line that tells the shell where the backend listens.
const announcePrefix = "drift-backend "

// frontend/build is the SvelteKit static build (`bun run build` in frontend/).
//
//go:embed all:frontend/build
var embeddedAssets embed.FS

// announcement is the JSON after announcePrefix.
type announcement struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func main() {
	bindingsDirectory := flag.String("bindings", "", "write the TypeScript bindings into this directory and exit")
	exitWithStdin := flag.Bool("exit-with-stdin", false, "quit when stdin closes (set by the Electron shell, so the backend dies with it)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()}))
	var err error
	if *bindingsDirectory != "" {
		err = writeBindings(*bindingsDirectory)
	} else {
		err = run(logger, *exitWithStdin)
	}
	if err != nil {
		logger.Error("desktop backend", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, exitWithStdin bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if exitWithStdin {
		ctx = cancelWhenStdinCloses(ctx)
	}

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	assets, err := fs.Sub(embeddedAssets, "frontend/build")
	if err != nil {
		return fmt.Errorf("frontend assets: %w", err)
	}

	token := rand.Text()
	baseURL := "http://" + listener.Addr().String()
	rpc := newHub(token, logger)
	app := NewApp(logger, baseURL, rpc.broadcast)
	server := &http.Server{
		Handler:           newServerHandler(app.proxyHandler(), rpc, app.neoworksCallbackHandler(), assets),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(listener) }()

	app.startup(ctx)
	rpc.serve(app)
	err = announce(os.Stdout, announcement{URL: baseURL, Token: token})
	if err != nil {
		return err
	}

	err = waitForStop(ctx, serveErrors)
	shutdown(server, app, logger)
	return err
}

func waitForStop(ctx context.Context, serveErrors <-chan error) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-serveErrors:
		return fmt.Errorf("server stopped: %w", err)
	}
}

func shutdown(server *http.Server, app *App, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Warn("server shutdown", "error", err)
	}
	app.shutdown()
}

func announce(writer io.Writer, message announcement) error {
	encoded, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("encode announcement: %w", err)
	}
	_, err = fmt.Fprintf(writer, "%s%s\n", announcePrefix, encoded)
	if err != nil {
		return fmt.Errorf("announce: %w", err)
	}
	return nil
}

// cancelWhenStdinCloses ends ctx when the parent closes our stdin, which also
// happens when the parent crashes.
func cancelWhenStdinCloses(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		_, _ = io.Copy(io.Discard, bufio.NewReader(os.Stdin))
		cancel()
	}()
	return ctx
}

// logLevel enables debug logs (e.g. every stream request) with DRIFT_LOG=debug.
func logLevel() slog.Level {
	if os.Getenv("DRIFT_LOG") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
