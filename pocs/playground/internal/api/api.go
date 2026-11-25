package api

import (
	"context"
	"encoding/json"
	"errors"
	"go/format"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/diegopacheco/golang-playground/pocs/playground/internal/complete"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/content"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/markdown"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/runner"
)

var ErrBadRequest = errors.New("bad request")

type codeRunner interface {
	Run(ctx context.Context, code string) (runner.Result, error)
}

type completer interface {
	Query(prefix string, limit int) []complete.Item
	Size() int
}

type Handler struct {
	runner    codeRunner
	completer completer
	logger    *slog.Logger
	goVersion string
}

type codeRequest struct {
	Code string `json:"code"`
}

type service struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type docView struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	HTML  string `json:"html"`
}

func New(r codeRunner, c completer, logger *slog.Logger, goVersion string) *Handler {
	return &Handler{runner: r, completer: c, logger: logger, goVersion: goVersion}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("POST /api/run", h.run)
	mux.HandleFunc("POST /api/format", h.format)
	mux.HandleFunc("GET /api/complete", h.complete)
	mux.HandleFunc("GET /api/docs", h.docs)
	mux.HandleFunc("GET /api/docs/{id}", h.doc)
	mux.HandleFunc("GET /api/samples", h.samples)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	docs, docsErr := content.Docs()
	samples, samplesErr := content.Samples()
	h.writeJSON(r.Context(), w, http.StatusOK, map[string]any{
		"status": "ok",
		"services": []service{
			{Name: "Go toolchain " + h.goVersion, Status: "ready"},
			{Name: "API server", Status: "ready"},
			{Name: "Completion index (" + strconv.Itoa(h.completer.Size()) + " symbols)", Status: "ready"},
			{Name: "Docs (" + strconv.Itoa(len(docs)) + " topics)", Status: status(docsErr)},
			{Name: "Samples (" + strconv.Itoa(len(samples)) + " programs)", Status: status(samplesErr)},
		},
	})
}

func (h *Handler) run(w http.ResponseWriter, r *http.Request) {
	req, err := decode(r)
	if err != nil {
		h.writeError(r.Context(), w, err)
		return
	}
	res, err := h.runner.Run(r.Context(), req.Code)
	if err != nil {
		h.writeError(r.Context(), w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "code run", "exit", res.ExitCode, "duration_ms", res.DurationMs, "timed_out", res.TimedOut)
	h.writeJSON(r.Context(), w, http.StatusOK, res)
}

func (h *Handler) format(w http.ResponseWriter, r *http.Request) {
	req, err := decode(r)
	if err != nil {
		h.writeError(r.Context(), w, err)
		return
	}
	out, err := format.Source([]byte(req.Code))
	if err != nil {
		h.writeJSON(r.Context(), w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	h.writeJSON(r.Context(), w, http.StatusOK, codeRequest{Code: string(out)})
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(r.Context(), w, http.StatusOK, h.completer.Query(r.URL.Query().Get("prefix"), 40))
}

func (h *Handler) docs(w http.ResponseWriter, r *http.Request) {
	docs, err := content.Docs()
	if err != nil {
		h.writeError(r.Context(), w, err)
		return
	}
	h.writeJSON(r.Context(), w, http.StatusOK, docs)
}

func (h *Handler) doc(w http.ResponseWriter, r *http.Request) {
	d, err := content.DocByID(r.PathValue("id"))
	if err != nil {
		h.writeError(r.Context(), w, err)
		return
	}
	h.writeJSON(r.Context(), w, http.StatusOK, docView{ID: d.ID, Title: d.Title, HTML: markdown.Render(d.Markdown)})
}

func (h *Handler) samples(w http.ResponseWriter, r *http.Request) {
	s, err := content.Samples()
	if err != nil {
		h.writeError(r.Context(), w, err)
		return
	}
	h.writeJSON(r.Context(), w, http.StatusOK, s)
}

func decode(r *http.Request) (codeRequest, error) {
	var req codeRequest
	if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 256*1024)).Decode(&req); err != nil {
		return req, errors.Join(ErrBadRequest, err)
	}
	return req, nil
}

func status(err error) string {
	if err != nil {
		return "error"
	}
	return "ready"
}

func (h *Handler) writeError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, content.ErrNotFound):
		h.writeJSON(ctx, w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, ErrBadRequest), errors.Is(err, runner.ErrEmptyCode):
		h.writeJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		h.logger.ErrorContext(ctx, "request failed", "err", err)
		h.writeJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func (h *Handler) writeJSON(ctx context.Context, w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		h.logger.ErrorContext(ctx, "encode response", "err", err)
	}
}
