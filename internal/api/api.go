package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/mqufflc/whodidthechores/internal/config"
	"github.com/mqufflc/whodidthechores/internal/html"
	"github.com/mqufflc/whodidthechores/internal/repository"
	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
)

type HTTPServer struct {
	repository *repository.Repository
	timezone   *time.Location
}

func New(repo *repository.Repository, conf config.Config) http.Handler {
	location, _ := time.LoadLocation(conf.TimeZone) //timezone already validated in config
	s := &HTTPServer{
		repository: repo,
		timezone:   location,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.notFound)
	mux.HandleFunc("/{$}", s.index)
	mux.HandleFunc("/chores", s.chores)
	mux.HandleFunc("/chores/{id}", s.editChore)
	mux.HandleFunc("/chores/new", s.createChore)
	mux.HandleFunc("/users", s.users)
	mux.HandleFunc("/users/new", s.createUser)
	mux.HandleFunc("/tasks", s.tasks)
	mux.HandleFunc("/tasks/new", s.createTask)
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
	if r.Method == "POST" {
		slog.Info("Received POST on /chores")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		default_duration, err := strconv.Atoi(r.FormValue("default_duration"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := h.repository.CreateChore(r.Context(), postgres.CreateChoreParams{Name: r.FormValue("name"), Description: r.FormValue("description"), DefaultDurationMn: int32(default_duration)}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
	h.viewChores(w, r)
}

func (h *HTTPServer) viewChores(w http.ResponseWriter, r *http.Request) {
	chores, err := h.repository.ListChores(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	html.Chores(chores).Render(r.Context(), w)
}

func (h *HTTPServer) createChore(w http.ResponseWriter, r *http.Request) {
	html.ChoreCreate().Render(r.Context(), w)
}

func (h *HTTPServer) editChore(w http.ResponseWriter, r *http.Request) {
	choreID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == "PUT" {
		default_duration, err := strconv.Atoi(r.FormValue("default_duration"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err = h.repository.UpdateChore(r.Context(), postgres.UpdateChoreParams{ID: int32(choreID), Name: r.FormValue("name"), Description: r.FormValue("description"), DefaultDurationMn: int32(default_duration)}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("internal server error: %v", err))
			return
		}
	}
	chore, err := h.repository.GetChore(r.Context(), int32(choreID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	html.ChoreEdit(chore).Render(r.Context(), w)
}

func (h *HTTPServer) users(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		slog.Info("Received POST on /users")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := h.repository.CreateUser(r.Context(), r.FormValue("name")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
	h.viewUsers(w, r)
}

func (h *HTTPServer) viewUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repository.ListUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	html.Users(users).Render(r.Context(), w)
}

func (h *HTTPServer) createUser(w http.ResponseWriter, r *http.Request) {
	html.UserCreate().Render(r.Context(), w)
}

func (h *HTTPServer) tasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		slog.Info("Received POST on /users")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		choreID, err := strconv.Atoi(r.FormValue("chore-id"))
		if err != nil {
			slog.Error(fmt.Sprintf("Unable to parse chore id: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userID, err := strconv.Atoi(r.FormValue("user-id"))
		if err != nil {
			slog.Error(fmt.Sprintf("Unable to parse user id: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		startedAt, err := time.ParseInLocation("2006-01-02T15:04", r.FormValue("start-time"), h.timezone)
		if err != nil {
			slog.Error(fmt.Sprintf("Unable to parse started time: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		duration, err := strconv.Atoi(r.FormValue("duration"))
		if err != nil {
			slog.Error(fmt.Sprintf("Unable to parse duration: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := h.repository.CreateTask(r.Context(), postgres.CreateTaskParams{
			ChoreID:     int32(choreID),
			UserID:      int32(userID),
			StartedAt:   startedAt,
			DurationMn:  int32(duration),
			Description: r.FormValue("description"),
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
	h.viewTasks(w, r)
}

func (h *HTTPServer) viewTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.repository.ListUsersTasks(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	html.Tasks(tasks, h.timezone).Render(r.Context(), w)
}

func (h *HTTPServer) createTask(w http.ResponseWriter, r *http.Request) {
	chores, err := h.repository.ListChores(r.Context())
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to list chores %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	users, err := h.repository.ListUsers(r.Context())
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to list users %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	html.TaskCreate(chores, users).Render(r.Context(), w)
}
