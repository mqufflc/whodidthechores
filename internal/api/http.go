package api

import (
	"log/slog"
	"net/http"

	"github.com/mqufflc/whodidthechores/internal/components"
	"github.com/mqufflc/whodidthechores/internal/repository"
	"github.com/mqufflc/whodidthechores/internal/utils"
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
	mux.HandleFunc("/login", s.login)
	mux.HandleFunc("/signup", s.signup)
	mux.HandleFunc("/chores", s.chores)
	mux.HandleFunc("/chores/new", s.createChoreForm)
	mux.HandleFunc("/", s.notFound)
	return mux
}

func (h *HTTPServerAPI) index(w http.ResponseWriter, r *http.Request) {
	components.Index().Render(r.Context(), w)
}

func (h *HTTPServerAPI) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		slog.Info("Received POST on /login")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			return
		}
		session, err := h.Repository.Login(repository.Credentials{Name: r.FormValue("username"), Password: r.FormValue("password")})
		if err != nil {
			w.WriteHeader(500)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: session.ID, HttpOnly: true, Expires: session.ExpiresAt})
		http.Redirect(w, r, "/", http.StatusFound)
	}
	components.Login().Render(r.Context(), w)
}

func (h *HTTPServerAPI) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		slog.Info("Received POST on /signup")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			slog.Error("Error while parsing form values : %w", err)
			return
		}
		session, err := h.Repository.SignUp(repository.Credentials{Name: r.FormValue("username"), Password: r.FormValue("password")})
		if err != nil {
			w.WriteHeader(500)
			slog.Error("Error while getting session after signup : %w", err)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: session.ID, HttpOnly: true, Expires: session.ExpiresAt})
		http.Redirect(w, r, "/", http.StatusFound)
	}
	components.SignUp().Render(r.Context(), w)
}

func (h *HTTPServerAPI) chores(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			return
		}
		if _, err := h.Repository.CreateChore(repository.ChoreParams{ID: utils.GenerateBase58ID(10), Name: r.FormValue("name"), Description: r.FormValue("description")}); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
	}
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

func (h *HTTPServerAPI) createChoreForm(w http.ResponseWriter, r *http.Request) {
	components.ChoreCreate().Render(r.Context(), w)
}
