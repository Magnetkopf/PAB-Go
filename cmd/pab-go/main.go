package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/app"
	"github.com/Magnetkopf/PAB-Go/internal/config"
	jsonstore "github.com/Magnetkopf/PAB-Go/internal/store/json"
	"github.com/Magnetkopf/PAB-Go/internal/telegram"
)

func main() {
	bind := flag.String("bind", "127.0.0.1:7212", "HTTP bind address (ip:port)")
	flag.Parse()

	cfg, err := config.LoadOrInit()
	if err != nil {
		panic(err)
	}
	store, err := jsonstore.New()
	if err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	botDone := make(chan struct{})
	go func() {
		defer close(botDone)
		telegram.NewClient(nil).Run(ctx, store)
	}()
	server := &http.Server{Addr: *bind, Handler: app.NewServer(cfg, store)}
	if err := serve(ctx, stop, server); err != nil {
		panic(err)
	}
	<-botDone
}

func serve(ctx context.Context, stop context.CancelFunc, server *http.Server) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()

	select {
	case err := <-serveErr:
		stop()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		// Restore normal signal handling so another Ctrl+C can stop a slow shutdown.
		stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			_ = server.Close()
		}
		err := <-serveErr
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return shutdownErr
	}
}
