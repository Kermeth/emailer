package main

import (
	"errors"
	"github.com/kermeth/emailer/internal"
	"log/slog"
	"net/http"
)

func main() {
	srv := internal.NewServer()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}
	slog.Info("server listening", "PORT", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed to start", "Error", err)
		panic(err)
	}
}
