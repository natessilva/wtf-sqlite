package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"sqlite/model"
	"sqlite/templates"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

//go:embed assets/*
var assetsFS embed.FS

type Handler struct {
	AuthService *AuthService
	UserService *UserService
	TaskService *TaskService
	UseTLS      bool
}

func NewHandler(authService *AuthService, userService *UserService, taskService *TaskService, useTLS bool) http.Handler {
	h := &Handler{
		AuthService: authService,
		UserService: userService,
		TaskService: taskService,
		UseTLS:      useTLS,
	}

	router := NewInstrumentedRouter()

	// Authenticated users will be redirected away from these routes
	router.GET("/login", requireNoAuth(h.handleGetLogin))
	router.POST("/login", requireNoAuth(h.handlePostLogin))
	router.GET("/signup", requireNoAuth(h.handleGetSignup))
	router.POST("/signup", requireNoAuth(h.handlePostSignup))

	// these routes are public.
	router.GET("/logout", h.handleLogout)
	router.GET("/", h.handleIndex)

	// Unauthenticated users will be redirected to login from these routes
	router.GET("/tasks", requireAuth(h.handleTasks))
	router.POST("/newTask", requireAuth(h.handlePostNewTask))
	router.POST("/insertTaskBefore", requireAuth(h.handleInsertBeforeTask))
	router.GET("/tasks/:id", requireAuth(h.handleGetTask))
	router.POST("/tasks/:id", requireAuth(h.handlePostEditTask))
	router.POST("/tasks/:id/delete", requireAuth(h.handleDeleteTask))

	mux := http.NewServeMux()
	mux.Handle("/", authService.Middleware(router))
	mux.Handle("/assets/", cache(http.FileServer(http.FS(assetsFS)), useTLS))

	router.NotFound = http.HandlerFunc(handleNotFound)
	router.PanicHandler = handleError

	return instrumentedHandler(gzipMiddleware(mux))
}

func cache(handler http.Handler, cache bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cache {
			w.Header().Add("Cache-Control", "max-age=300, public, must-revalidate, no-transform")
		}
		handler.ServeHTTP(w, r)
	}
}

func requireNoAuth(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		userId := UserFromContext(r.Context()).ID
		if userId != 0 {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		handle(w, r, p)
	}
}

func requireAuth(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		userId := UserFromContext(r.Context()).ID
		if userId == 0 {
			http.Redirect(w, r, fmt.Sprintf("/login?next=%s", r.URL.Path), http.StatusSeeOther)
			return
		}
		handle(w, r, p)
	}
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := UserFromContext(r.Context()).ID
	if userId == 0 {
		templates.IndexNoAuth().Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handler) handleGetLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	templates.AuthForm(true, "", "", r.FormValue("next")).Render(r.Context(), w)
}

func (h *Handler) handlePostLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	output, err := h.AuthService.Login(r.Context(), AuthInput{
		UserName: userName,
		Password: password,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}
	if !output.OK {
		w.WriteHeader(http.StatusUnauthorized)
		templates.AuthForm(true, "Invalid email and/or password", userName, r.FormValue("next")).Render(r.Context(), w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    output.Token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().AddDate(0, 0, 30),
		Secure:   h.UseTLS,
	})
	redirect := r.FormValue("next")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h *Handler) handleGetSignup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	templates.AuthForm(false, "", "", r.FormValue("next")).Render(r.Context(), w)
}

func (h *Handler) handlePostSignup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	next := r.FormValue("next")
	if userName == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		templates.AuthForm(false, "Missing required values", userName, next).Render(r.Context(), w)
		return
	}
	output, err := h.AuthService.Signup(r.Context(), AuthInput{
		UserName: userName,
		Password: password,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}
	if !output.OK {
		w.WriteHeader(http.StatusUnauthorized)
		templates.AuthForm(false, "Username already claimed", userName, next).Render(r.Context(), w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    output.Token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().AddDate(0, 0, 30),
		Secure:   h.UseTLS,
	})
	redirect := r.FormValue("next")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	cookie, err := r.Cookie("token")
	if err == nil {
		// clear and expire the cookie
		cookie.Value = ""
		cookie.Expires = time.Unix(0, 0)
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tasks, err := h.TaskService.List(r.Context(), sql.NullInt64{})
	if err != nil {
		handleError(w, r, err)
		return
	}
	templates.TaskListPage(tasks, sql.NullInt64{}).Render(r.Context(), w)
}

func (h *Handler) handlePostNewTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	parent := r.FormValue("parentID")
	var parentID sql.NullInt64
	if parent != "" {
		parentID.Int64, _ = strconv.ParseInt(parent, 10, 64)
		parentID.Valid = true
	}
	task, err := h.TaskService.Create(r.Context(), title, description, parentID)
	if err != nil {
		handleError(w, r, err)
		return
	}
	templates.TaskItem(task, true).Render(r.Context(), w)
}

func (h *Handler) handleGetTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, _ := strconv.ParseInt(p.ByName("id"), 10, 64)
	task, err := h.TaskService.Get(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFound(true).Render(r.Context(), w)
		}
		handleError(w, r, err)
		return
	}
	children, err := h.TaskService.List(r.Context(), sql.NullInt64{Int64: task.ID, Valid: true})
	templates.Task(task, children).Render(r.Context(), w)
}

func (h *Handler) handlePostEditTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, _ := strconv.ParseInt(p.ByName("id"), 10, 64)
	title := r.FormValue("title")
	description := r.FormValue("description")
	err := h.TaskService.Update(r.Context(), model.Task{
		ID:          id,
		Title:       title,
		Description: description,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFound(true).Render(r.Context(), w)
			return
		}
		handleError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/tasks"), http.StatusFound)
}

func (h *Handler) handleDeleteTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, _ := strconv.ParseInt(p.ByName("id"), 10, 64)

	err := h.TaskService.Delete(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFound(true).Render(r.Context(), w)
			return
		}
		handleError(w, r, err)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handler) handleInsertBeforeTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	idToInsert, _ := strconv.ParseInt(r.FormValue("idToInsert"), 10, 64)
	targetStr := r.FormValue("target")
	target := sql.NullInt64{}
	if targetStr != "" {
		target.Int64, _ = strconv.ParseInt(targetStr, 10, 64)
		target.Valid = true
	}
	parent := r.FormValue("parentID")
	parentID := sql.NullInt64{}
	if parent != "" {
		parentID.Int64, _ = strconv.ParseInt(parent, 10, 64)
		parentID.Valid = true
	}

	err := h.TaskService.InsertBefore(r.Context(), idToInsert, target, parentID)
	if err != nil {
		handleError(w, r, err)
		return
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err interface{}) {
	ctx := r.Context()
	w.WriteHeader(http.StatusInternalServerError)
	templates.Error(UserFromContext(ctx).ID != 0).Render(ctx, w)
	path := "unknown"
	if rec, ok := w.(*instrumentedResponseWriter); ok {
		path = rec.path
	}
	fmt.Printf("path:%s error: %v\n", path, err)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.WriteHeader(http.StatusNotFound)
	templates.NotFound(UserFromContext(ctx).ID != 0).Render(ctx, w)
}
