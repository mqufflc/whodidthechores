package api

import (
	"net/http"

	"github.com/mqufflc/whodidthechores/internal/html"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

type HTTPServer struct {
	repository *repository.Repository
}

func New(repo *repository.Repository) http.Handler {
	s := &HTTPServer{
		repository: repo,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.notFound)
	mux.HandleFunc("/{$}", s.index)
	mux.HandleFunc("/chores", s.chores)
	return mux
}

func (h *HTTPServer) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
	html.NotFound().Render(r.Context(), w)
}

func (h *HTTPServer) index(w http.ResponseWriter, r *http.Request) {
	html.Index().Render(r.Context(), w)
}

func (h *HTTPServer) chores(w http.ResponseWriter, r *http.Request) {
	h.viewChores(w, r)
}

func (h *HTTPServer) viewChores(w http.ResponseWriter, r *http.Request) {
	chores, err := h.repository.ListChores(r.Context())
	if err != nil {
		w.WriteHeader(500)
	}
	html.Chores(chores).Render(r.Context(), w)
}
