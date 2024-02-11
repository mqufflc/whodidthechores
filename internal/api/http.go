package api

import (
	"net/http"

	"github.com/mqufflc/whodidthechores/internal/components"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

type DefaultHandler struct {
	Repository *repository.Service
}

func New(repo *repository.Service) *DefaultHandler {
	return &DefaultHandler{
		Repository: repo,
	}
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Get(w, r)
}

func (h *DefaultHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.View(w, r)
}

func (h *DefaultHandler) View(w http.ResponseWriter, r *http.Request) {
	chores, err := h.Repository.ListChores()
	if err != nil {
		w.WriteHeader(500)
	}
	components.Chores(*chores).Render(r.Context(), w)
}
