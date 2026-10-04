package tasks

import (
	"net/http"

	"github.com/leafsource/gosilite/httpx"
)

// Handler exposes Service over HTTP/JSON.
type Handler struct {
	svc *Service
}

// NewHandler builds the HTTP adapter.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers task REST endpoints on mux.
// Pattern syntax follows Go 1.22+ net/http.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/tasks", h.create)
	mux.HandleFunc("GET /api/tasks", h.list)
	mux.HandleFunc("GET /api/tasks/{id}", h.get)
	mux.HandleFunc("PATCH /api/tasks/{id}", h.update)
	mux.HandleFunc("DELETE /api/tasks/{id}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateTaskRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.RespondErr(w, err)
		return
	}
	task, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpx.RespondErr(w, err)
		return
	}
	httpx.RespondCreated(w, task)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := ListTasksQuery{Status: r.URL.Query().Get("status")}
	tasks, err := h.svc.List(r.Context(), q)
	if err != nil {
		httpx.RespondErr(w, err)
		return
	}
	httpx.RespondOK(w, tasks)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	task, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.RespondErr(w, err)
		return
	}
	httpx.RespondOK(w, task)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateTaskRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.RespondErr(w, err)
		return
	}
	task, err := h.svc.Update(r.Context(), r.PathValue("id"), in)
	if err != nil {
		httpx.RespondErr(w, err)
		return
	}
	httpx.RespondOK(w, task)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.RespondErr(w, err)
		return
	}
	httpx.RespondNoContent(w)
}
