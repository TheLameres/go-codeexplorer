package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-codeexplorer/dtos"
	"go-codeexplorer/models"
	"io"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, dtos.ErrorResponse{Error: msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	// в теле не должно быть второго JSON-значения
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single JSON object")
	}
	return nil
}

func (h *TodoHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, models.ErrNotFound):
		writeError(w, http.StatusNotFound, "todo not found")
	case errors.Is(err, r.Context().Err()):
		// клиент ушёл или истёк таймаут — отвечать уже некому
		h.log.Warn("request cancelled", "path", r.URL.Path, "err", err)
	default:
		// детали ошибки — только в лог, клиенту ничего лишнего
		h.log.Error("internal error", "path", r.URL.Path, "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
