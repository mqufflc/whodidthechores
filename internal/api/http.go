package api

import (
	"net/http"

	"github.com/mqufflc/whodidthechores/internal/components"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

type HTTPServerAPI struct {
	Repository *repository.Service
}

func New(repo *repository.Service) http.Handler {
	s := &HTTPServerAPI{
		Repository: repo,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", s.index)
	mux.HandleFunc("/chores/{$}", s.chores)
	mux.HandleFunc("/", s.notFound)
	return mux
}

func (h *HTTPServerAPI) index(w http.ResponseWriter, r *http.Request) {
	components.Index().Render(r.Context(), w)
}

func (h *HTTPServerAPI) chores(w http.ResponseWriter, r *http.Request) {
	h.viewChores(w, r)
}

func (h *HTTPServerAPI) viewChores(w http.ResponseWriter, r *http.Request) {
	chores, err := h.Repository.ListChores()
	if err != nil {
		w.WriteHeader(500)
	}
	components.Chores(*chores).Render(r.Context(), w)
}

func (h *HTTPServerAPI) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	components.NotFound().Render(r.Context(), w)
}
