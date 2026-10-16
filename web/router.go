package web

import (
	"go-codeexplorer/dtos"
	"log/slog"
	"net/http"
)

func NewRouter(todo *TodoHandler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, dtos.ErrorResponse{Error: "Route not found"})
	})
	todo.Register(mux)

	var h http.Handler = mux
	h = Logging(log)(h)
	h = Recover(log)(h)
	return h
}
