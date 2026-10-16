package web

import (
	"context"
	"go-codeexplorer/dtos"
	"log/slog"
	"net/http"
	"strconv"
)

type TodoService interface {
	Add(ctx context.Context, req dtos.CreateNewTodoEntryRequest) (dtos.CreateNewTodoEntryResponse, error)
	GetByID(ctx context.Context, id int64) (dtos.TodoEntryResponse, error)
	List(ctx context.Context) ([]dtos.TodoEntryResponse, error)
	MarkDone(ctx context.Context, id int64) error
}

type TodoHandler struct {
	svc TodoService
	log *slog.Logger
}

func NewTodoHandler(svc TodoService, log *slog.Logger) *TodoHandler {
	return &TodoHandler{svc: svc, log: log}
}

func (h *TodoHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /todos", h.create)
	mux.HandleFunc("GET /todos", h.list)
	mux.HandleFunc("GET /todos/{id}", h.get)
	mux.HandleFunc("POST /todos/{id}/done", h.markDone)
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateNewTodoEntryRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg := req.Validate(); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}

	todo, err := h.svc.Add(r.Context(), req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	w.Header().Set("Location", "/todos/"+strconv.FormatInt(todo.Id, 10))
	writeJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	todo, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	todos, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	// make, а не nil-слайс: иначе JSON будет null вместо []
	res := make([]dtos.TodoEntryResponse, 0, len(todos))
	res = append(res, todos...)
	writeJSON(w, http.StatusOK, res)
}

func (h *TodoHandler) markDone(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.MarkDone(r.Context(), id); err != nil {
		h.handleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
