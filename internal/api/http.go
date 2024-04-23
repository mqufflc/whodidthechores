package api

import (
	"log/slog"
	"net/http"

	"github.com/mqufflc/whodidthechores/internal/components"
	"github.com/mqufflc/whodidthechores/internal/middleware"
	"github.com/mqufflc/whodidthechores/internal/repository"
	"github.com/mqufflc/whodidthechores/internal/utils"
)

type HTTPServerAPI struct {
	repository *repository.Service
}

func New(repo *repository.Service) http.Handler {
	s := &HTTPServerAPI{
		repository: repo,
	}
	authMiddlewareFactory := middleware.NewAuthMiddlewareFactory(s.repository)
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", s.index)
	mux.HandleFunc("/login", s.login)
	mux.HandleFunc("/signup", s.signup)
	mux.Handle("/chores", authMiddlewareFactory.EnsureAuth(s.chores))
	mux.HandleFunc("/chores/new", s.createChoreForm)
	mux.HandleFunc("/", s.notFound)
	return mux
}

func (h *HTTPServerAPI) index(w http.ResponseWriter, r *http.Request) {
	components.Index().Render(r.Context(), w)
}

func (h *HTTPServerAPI) login(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("redirect")
	if r.Method == "POST" {
		slog.Info("Received POST on /login")
		if err := r.ParseForm(); err != nil {
			http.Error(w, "unable to parse form", http.StatusBadRequest)
			return
		}
		session, err := h.repository.Login(repository.Credentials{Name: r.FormValue("username"), Password: r.FormValue("password")})
		if err != nil {
			w.WriteHeader(200)
			components.Login(redirect).Render(r.Context(), w)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: session.ID.String(), HttpOnly: true, Expires: session.ExpiresAt, Secure: true, SameSite: http.SameSiteStrictMode})
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
	if r.Method == "GET" {

	}
	components.Login(redirect).Render(r.Context(), w)
}

func (h *HTTPServerAPI) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		slog.Info("Received POST on /signup")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			slog.Error("Error while parsing form values : %w", err)
			return
		}
		session, err := h.repository.SignUp(repository.Credentials{Name: r.FormValue("username"), Password: r.FormValue("password")})
		if err != nil {
			w.WriteHeader(500)
			slog.Error("Error while getting session after signup : %w", err)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: session.ID.String(), HttpOnly: true, Expires: session.ExpiresAt, Secure: false})
		http.Redirect(w, r, "/", http.StatusFound)
	}
	components.SignUp().Render(r.Context(), w)
}

func (h *HTTPServerAPI) chores(w http.ResponseWriter, r *http.Request, user *repository.User) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			return
		}
		if _, err := h.repository.CreateChore(repository.ChoreParams{ID: utils.GenerateBase58ID(10), Name: r.FormValue("name"), Description: r.FormValue("description")}); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
	}
	h.viewChores(w, r)
}

func (h *HTTPServerAPI) viewChores(w http.ResponseWriter, r *http.Request) {
	chores, err := h.repository.ListChores()
	if err != nil {
		w.WriteHeader(500)
	}
	components.Chores(*chores).Render(r.Context(), w)
}

func (h *HTTPServerAPI) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
	components.NotFound().Render(r.Context(), w)
}

func (h *HTTPServerAPI) createChoreForm(w http.ResponseWriter, r *http.Request) {
	components.ChoreCreate().Render(r.Context(), w)
}
