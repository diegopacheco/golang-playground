package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/node"
)

const maxBody = 1 << 20

type Store interface {
	Get(key string) (string, bool)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	Join(ctx context.Context, req node.JoinRequest) error
	Status() node.Status
	IsLeader() bool
	LeaderHTTP() (string, bool)
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Server struct {
	store Store
}

func New(store Store) *Server {
	return &Server{store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /kv/{key}", s.get)
	mux.HandleFunc("PUT /kv/{key}", s.leaderOnly(s.put))
	mux.HandleFunc("DELETE /kv/{key}", s.leaderOnly(s.del))
	mux.HandleFunc("GET /status", s.status)
	mux.HandleFunc("POST /join", s.leaderOnly(s.join))
	return mux
}

func (s *Server) leaderOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.store.IsLeader() {
			next(w, r)
			return
		}
		s.redirect(w, r)
	}
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	addr, ok := s.store.LeaderHTTP()
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "no leader known")
		return
	}
	http.Redirect(w, r, "http://"+addr+r.URL.RequestURI(), http.StatusTemporaryRedirect)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	v, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "key not found")
		return
	}
	writeJSON(w, http.StatusOK, KV{Key: key, Value: v})
}

func (s *Server) put(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}
	if err := s.store.Set(r.Context(), key, string(body)); err != nil {
		s.writeFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, KV{Key: key, Value: string(body)})
}

func (s *Server) del(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Delete(r.Context(), r.PathValue("key")); err != nil {
		s.writeFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) status(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Status())
}

func (s *Server) join(w http.ResponseWriter, r *http.Request) {
	var req node.JoinRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid join body")
		return
	}
	if err := s.store.Join(r.Context(), req); err != nil {
		s.writeFailure(w, r, err)
		return
	}
	logx.FromContext(r.Context()).Info("node joined", "id", req.ID, "raft_addr", req.RaftAddr, "http_addr", req.HTTPAddr)
	writeJSON(w, http.StatusOK, req)
}

func (s *Server) writeFailure(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, node.ErrNotLeader):
		s.redirect(w, r)
	case errors.Is(err, node.ErrBadRequest):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		logx.FromContext(r.Context()).Error("request failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
