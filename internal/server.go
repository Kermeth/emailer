package internal

import (
	"github.com/kermeth/emailer/internal/gmail"
	"github.com/kermeth/emailer/internal/health"
	"github.com/kermeth/emailer/internal/send"
	"net/http"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)
	var handler http.Handler = mux
	return handler
}

func addRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", health.Handler)
	mux.HandleFunc("POST /smtp/send", send.Handler)
	mux.HandleFunc("POST /api/gmail/send", gmail.Handler)
}
