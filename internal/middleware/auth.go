package middleware

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

type AuthenticatedHandler func(http.ResponseWriter, *http.Request, *repository.User)

type AuthMiddleware struct {
	repository *repository.Service
	handler    AuthenticatedHandler
}

type AuthMiddlewareFactory struct {
	repository *repository.Service
}

func (m *AuthMiddleware) getSession(r *http.Request) (*repository.Session, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return &repository.Session{}, err
	}
	uuid, err := uuid.Parse(cookie.Value)
	if err != nil {
		return &repository.Session{}, err
	}
	session, err := m.repository.GetSession(uuid)
	if err != nil {
		return &repository.Session{}, err
	}

	if session == nil {
		return &repository.Session{}, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return &repository.Session{}, errors.New("session expired")
	}

	m.repository.UseSession(session.ID)

	return session, nil
}

func (authMiddleware AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := authMiddleware.getSession(r)
	if err != nil {
		if r.URL.Path == "/login" {
			authMiddleware.handler(w, r, nil)
			return
		}
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		redirect := r.URL.Path + query
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "", HttpOnly: true, Expires: time.Unix(0, 0), MaxAge: -1, Secure: true, Path: "/"})
		http.Redirect(w, r, "/login?redirect="+url.PathEscape(redirect), http.StatusTemporaryRedirect)
		return
	}
	authMiddleware.handler(w, r, session.User)
}

func (authMiddlewareFactory AuthMiddlewareFactory) EnsureAuth(handlerWrapped AuthenticatedHandler) *AuthMiddleware {
	return &AuthMiddleware{
		repository: authMiddlewareFactory.repository,
		handler:    handlerWrapped,
	}
}

func NewAuthMiddlewareFactory(repository *repository.Service) *AuthMiddlewareFactory {
	return &AuthMiddlewareFactory{
		repository: repository,
	}
}
