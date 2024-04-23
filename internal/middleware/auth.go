package middleware

import (
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

func (m *AuthMiddleware) IsAuthenticated(r *http.Request) (*repository.User, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return &repository.User{}, err
	}
	uuid, err := uuid.Parse(cookie.Value)
	if err != nil {
		return &repository.User{}, err
	}
	session, err := m.repository.GetSession(uuid)
	if err != nil {
		return &repository.User{}, err
	}

	if time.Now().After(session.ExpiresAt) {
		return &repository.User{}, err
	}

	return session.User, nil
}

func (authMiddleware AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := authMiddleware.IsAuthenticated(r)
	if err != nil {
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		redirect := r.URL.Path + query
		http.Redirect(w, r, "/login?redirect="+url.PathEscape(redirect), http.StatusTemporaryRedirect)
		return
	}
	authMiddleware.handler(w, r, user)
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
