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
	mux.HandleFunc("/static/{fileName}", serveStatic)
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

func serveStatic(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("fileName")
	if fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p, err := html.EmbedStatic.ReadFile(fmt.Sprintf("static/%s", fileName))
	if err != nil {
		slog.Error(fmt.Sprintf("unable to read static file: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if strings.HasSuffix(fileName, ".js") {
		w.Header().Set("Content-Type", "text/javascript")
	}
	if strings.HasSuffix(fileName, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	w.Write(p)
}

func (h *HTTPServer) notFound(w http.ResponseWriter, r *http.Request) {
	html.NotFound().Render(r.Context(), w)
}

func (h *HTTPServer) index(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	fromQuery := queries.Get("from")
	toQuery := queries.Get("to")
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, h.timezone)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	var from, to time.Time
	var err error
	if fromQuery != "" {
		from, err = time.ParseInLocation("2006-01-02T15:04", fromQuery, h.timezone)
		if err != nil {
			slog.Warn(fmt.Sprintf("Unable to parse 'from': %s", fromQuery))
			from = firstOfMonth
		}
	} else {
		from = firstOfMonth
	}
	if toQuery != "" {
		to, err = time.ParseInLocation("2006-01-02T15:04", toQuery, h.timezone)
		if err != nil {
			slog.Warn(fmt.Sprintf("Unable to parse 'to': %s", toQuery))
			to = lastOfMonth
		}
	} else {
		to = lastOfMonth
	}
	report, err := h.repository.GetChoreReport(r.Context(), from, to)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	chart := html.CreateBarChart(report)
	html.Index(chart, h.timezone, from, to).Render(r.Context(), w)
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
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Warn(fmt.Sprintf("unable to parse form: %v", err))
			return
		}
		taskParams := repository.TaskParams{
			ID:          uuid.UUID{},
			ChoreID:     r.FormValue("chore-id"),
			UserID:      r.FormValue("user-id"),
			StartedAt:   r.FormValue("start-time"),
			DurationMn:  r.FormValue("duration"),
			Description: r.FormValue("description"),
		}
		taskParamsValidated, err := h.repository.ValidateTask(r.Context(), &taskParams, *h.timezone)
		if err != nil {
			if errors.Is(err, repository.ErrValidation) {
				w.WriteHeader(http.StatusOK)
				html.TaskCreate(taskParams, chores, users).Render(r.Context(), w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.repository.CreateTask(r.Context(), taskParamsValidated)
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	taskParams := repository.TaskParams{
		ID:          uuid.UUID{},
		UserID:      strconv.FormatInt(int64(users[0].ID), 10),
		ChoreID:     strconv.FormatInt(int64(chores[0].ID), 10),
		StartedAt:   time.Now().In(h.timezone).Format("2006-01-02T15:04"),
		DurationMn:  "",
		Description: "",
	}
	html.TaskCreate(taskParams, chores, users).Render(r.Context(), w)
}

func (h *HTTPServer) editTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
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
	if r.Method == "PUT" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Warn(fmt.Sprintf("unable to parse form: %v", err))
			return
		}
		taskParams := repository.TaskParams{
			ID:          task.ID,
			ChoreID:     r.FormValue("chore-id"),
			UserID:      r.FormValue("user-id"),
			StartedAt:   r.FormValue("start-time"),
			DurationMn:  r.FormValue("duration"),
			Description: r.FormValue("description"),
		}
		taskParamsValidated, err := h.repository.ValidateTask(r.Context(), &taskParams, *h.timezone)
		if err != nil {
			if errors.Is(err, repository.ErrValidation) {
				w.WriteHeader(http.StatusOK)
				html.TaskEdit(taskParams, chores, users).Render(r.Context(), w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		task, err = h.repository.UpdateTask(r.Context(), task.ID, postgres.CreateTaskParams{
			ChoreID:     taskParamsValidated.ChoreID,
			UserID:      taskParamsValidated.UserID,
			StartedAt:   taskParamsValidated.StartedAt,
			DurationMn:  taskParamsValidated.DurationMn,
			Description: taskParamsValidated.Description,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("unable to edit task: %v", err))
			return
		}
	}
	taskParams := repository.TaskParams{
		ID:          task.ID,
		UserID:      strconv.FormatInt(int64(task.UserID), 10),
		ChoreID:     strconv.FormatInt(int64(task.ChoreID), 10),
		StartedAt:   task.StartedAt.In(h.timezone).Format("2006-01-02T15:04"),
		DurationMn:  strconv.FormatInt(int64(task.DurationMn), 10),
		Description: task.Description,
	}
	html.TaskEdit(taskParams, chores, users).Render(r.Context(), w)
}
