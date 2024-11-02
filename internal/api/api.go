package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
	mux.HandleFunc("/users/{id}", s.editUser)
	mux.HandleFunc("/users/new", s.createUser)
	mux.HandleFunc("/tasks", s.tasks)
	mux.HandleFunc("/tasks/{id}", s.editTask)
	mux.HandleFunc("/tasks/new", s.createTask)
	return mux
}

func (h *HTTPServer) notFound(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("unable to list chores: %v", err))
		return
	}
	html.Chores(chores).Render(r.Context(), w)
}

func (h *HTTPServer) createChore(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Warn(fmt.Sprintf("unable to parse form: %v", err))
			return
		}
		choreParams := repository.ChoreParams{ID: -1, Name: strings.TrimSpace(r.FormValue("name")), Description: strings.TrimSpace(r.FormValue("description")), DefaultDurationMn: r.FormValue("default_duration")}
		choreParamsValidated, err := h.repository.ValidateChore(r.Context(), &choreParams)
		if err != nil {
			if errors.Is(err, repository.ErrValidation) {
				w.WriteHeader(http.StatusOK)
				html.ChoreCreate(choreParams).Render(r.Context(), w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to validate chore: %v", err))
			return
		}
		if _, err := h.repository.CreateChore(r.Context(), choreParamsValidated); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to create chore: %v", err))
			return
		}
		http.Redirect(w, r, "/chores", http.StatusSeeOther)
		return
	}
	html.ChoreCreate(repository.ChoreParams{}).Render(r.Context(), w)
}

func (h *HTTPServer) editChore(w http.ResponseWriter, r *http.Request) {
	choreID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chore, err := h.repository.GetChore(r.Context(), int32(choreID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		html.NotFound().Render(r.Context(), w)
		return
	}
	if r.Method == "PUT" {
		choreParams := repository.ChoreParams{ID: chore.ID, Name: strings.TrimSpace(r.FormValue("name")), Description: strings.TrimSpace(r.FormValue("description")), DefaultDurationMn: r.FormValue("default_duration")}
		choreParamsValidated, err := h.repository.ValidateChore(r.Context(), &choreParams)
		if err != nil {
			if errors.Is(err, repository.ErrValidation) {
				w.WriteHeader(http.StatusOK)
				html.ChoreEdit(choreParams).Render(r.Context(), w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to validate chore: %v", err))
			return
		}
		chore, err = h.repository.UpdateChore(r.Context(), chore.ID, choreParamsValidated)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to edit chore: %v", err))
			return
		}
	}
	if r.Method == "DELETE" {
		err = h.repository.DeleteChore(r.Context(), chore.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to delete chore: %v", err))
			return
		}
		w.Header().Add("HX-Location", "/chores")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	choreParams := repository.ChoreParams{
		ID:                chore.ID,
		Name:              chore.Name,
		Description:       chore.Description,
		DefaultDurationMn: strconv.FormatInt(int64(chore.DefaultDurationMn), 10),
	}
	html.ChoreEdit(choreParams).Render(r.Context(), w)
}

func (h *HTTPServer) users(w http.ResponseWriter, r *http.Request) {
	h.viewUsers(w, r)
}

func (h *HTTPServer) viewUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repository.ListUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("unable to get users: %v", err))
		return
	}
	html.Users(users).Render(r.Context(), w)
}

func (h *HTTPServer) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("user create parsing: %v", err))
			return
		}
		userParams := repository.UserParams{
			ID:   -1,
			Name: r.FormValue("name"),
		}
		validatedName, err := h.repository.ValidateUser(r.Context(), &userParams)
		if err != nil {
			if errors.Is(err, repository.ErrValidation) {
				w.WriteHeader(http.StatusOK)
				html.UserCreate(userParams).Render(r.Context(), w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := h.repository.CreateUser(r.Context(), validatedName); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("user create error: %v", err))
			return
		}
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}
	html.UserCreate(repository.UserParams{}).Render(r.Context(), w)
}

func (h *HTTPServer) editUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.repository.GetUser(r.Context(), int32(userID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		html.NotFound().Render(r.Context(), w)
		return
	}
	if r.Method == "PUT" {
		userParams := repository.UserParams{
			ID:   user.ID,
			Name: strings.TrimSpace(r.FormValue("name")),
		}
		validatedName, err := h.repository.ValidateUser(r.Context(), &userParams)
		if err != nil {
			if errors.Is(err, repository.ErrValidation) {
				w.WriteHeader(http.StatusOK)
				html.UserEdit(userParams).Render(r.Context(), w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to validate user: %v", err))
			return
		}
		user, err = h.repository.UpdateUser(r.Context(), user.ID, validatedName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to edit user: %v", err))
			return
		}
	}
	if r.Method == "DELETE" {
		err = h.repository.DeleteUser(r.Context(), user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("user delete error: %v", err))
			return
		}
		w.Header().Add("HX-Location", "/users")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userParams := repository.UserParams{
		ID:   user.ID,
		Name: user.Name,
	}
	html.UserEdit(userParams).Render(r.Context(), w)
}

func (h *HTTPServer) tasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		slog.Info("Received POST on /tasks")
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("internal server error: %v", err))
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

func (h *HTTPServer) editTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == "PUT" {
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
		if _, err := h.repository.UpdateTask(r.Context(), postgres.UpdateTaskParams{
			ID:          taskID,
			ChoreID:     int32(choreID),
			UserID:      int32(userID),
			StartedAt:   startedAt,
			DurationMn:  int32(duration),
			Description: r.FormValue("description"),
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	task, err := h.repository.GetTask(r.Context(), taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method == "DELETE" {
		err = h.repository.DeleteTask(r.Context(), task.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("HX-Location", "/tasks")
		w.WriteHeader(http.StatusNoContent)
		return
	}
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
	html.TaskEdit(task, chores, users, h.timezone).Render(r.Context(), w)
}
