package sqlite

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
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
	DialService *DialService
	UseTLS      bool
}

func NewHandler(authService *AuthService, userService *UserService, dialService *DialService, useTLS bool) http.Handler {
	mux := http.NewServeMux()
	h := &Handler{
		AuthService: authService,
		UserService: userService,
		DialService: dialService,
		UseTLS:      useTLS,
	}

	router := NewInstrumentedRouter()

	// if you are already authenticated, none of these routes make
	// sense
	router.GET("/login", requireNoAuth(h.handleGetLogin))
	router.POST("/login", requireNoAuth(h.handlePostLogin))
	router.GET("/signup", requireNoAuth(h.handleGetSignup))
	router.POST("/signup", requireNoAuth(h.handlePostSignup))

	// these routes are public.
	router.GET("/logout", h.handleLogout)
	router.GET("/", h.handleIndex)

	// these routes required an authenticated user
	router.GET("/dials", requireAuth(h.handleDials))
	router.GET("/newDial", requireAuth(h.handleGetNewDials))
	router.POST("/newDial", requireAuth(h.handlePostNewDials))
	router.GET("/dials/:id", requireAuth(h.handleGetDial))
	router.GET("/dials/:id/edit", requireAuth(h.handleGetEditDial))
	router.POST("/dials/:id/edit", requireAuth(h.handlePostEditDial))
	router.PATCH("/dials/:id", requireAuth(h.handlePatchDial))
	router.POST("/dials/:id/delete", requireAuth(h.handleDeleteDial))

	mux.Handle("/", authService.Middleware(router))
	mux.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	router.NotFound = http.HandlerFunc(handleNotFound)
	router.PanicHandler = handleError

	return instrumentedHandler(mux)
}

func requireNoAuth(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		userId := UserFromFromContext(r.Context()).UserID
		if userId != 0 {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		handle(w, r, p)
	}
}

func requireAuth(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		userId := UserFromFromContext(r.Context()).UserID
		if userId == 0 {
			http.Redirect(w, r, fmt.Sprintf("/login?next=%s", r.URL.Path), http.StatusSeeOther)
			return
		}
		handle(w, r, p)
	}
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := UserFromFromContext(r.Context()).UserID
	if userId == 0 {
		templates.IndexNoAuth().Render(r.Context(), w)
		return
	}
	user, err := h.UserService.Get(r.Context())
	if err != nil {
		handleError(w, r, err)
		return
	}
	templates.Index(user.UserName).Render(r.Context(), w)
}

func (h *Handler) handleGetLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	templates.Login("", "", r.FormValue("next")).Render(r.Context(), w)
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
		templates.Login("Invalid email and/or password", userName, r.FormValue("next")).Render(r.Context(), w)
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
	templates.Signup("", "").Render(r.Context(), w)
}

func (h *Handler) handlePostSignup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	if userName == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		templates.Signup("Missing required values", userName).Render(r.Context(), w)
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
		templates.Signup("Username already claimed", userName).Render(r.Context(), w)
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

func (h *Handler) handleDials(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	dials, err := h.DialService.List(r.Context())
	if err != nil {
		handleError(w, r, err)
		return
	}
	templates.Dials(dials).Render(r.Context(), w)
}

func (h *Handler) handleGetNewDials(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	templates.DialForm("").Render(r.Context(), w)
}

func (h *Handler) handlePostNewDials(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := r.FormValue("name")
	id, err := h.DialService.Create(r.Context(), name)
	if err != nil {
		handleError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/dials/%d", id), http.StatusFound)
}

func (h *Handler) handleGetDial(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		handleError(w, r, err)
		return
	}
	dial, err := h.DialService.Get(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFound(true).Render(r.Context(), w)
			return
		}
		handleError(w, r, err)
		return
	}
	templates.Dial(dial).Render(r.Context(), w)
}

func (h *Handler) handleGetEditDial(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		handleError(w, r, err)
		return
	}
	dial, err := h.DialService.Get(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFound(true).Render(r.Context(), w)
		}
		handleError(w, r, err)
		return
	}
	templates.DialForm(dial.Name).Render(r.Context(), w)
}

func (h *Handler) handlePostEditDial(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		handleError(w, r, err)
		return
	}
	name := r.FormValue("name")
	err = h.DialService.Update(r.Context(), UpdateDial{
		ID:   id,
		Name: name,
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
	http.Redirect(w, r, fmt.Sprintf("/dials/%d", id), http.StatusFound)
}

type PatchDial struct {
	Value int64 `json:"value"`
}

func (h *Handler) handlePatchDial(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		handleError(w, r, err)
		return
	}

	var patch PatchDial
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&patch)
	if err != nil {
		handleError(w, r, err)
		return
	}
	err = h.DialService.SetValue(r.Context(), SetDialValue{
		ID:    id,
		Value: patch.Value,
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
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleDeleteDial(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		handleError(w, r, err)
		return
	}

	err = h.DialService.Delete(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFound(true).Render(r.Context(), w)
			return
		}
		handleError(w, r, err)
		return
	}
	http.Redirect(w, r, "/dials", http.StatusSeeOther)
}

func handleError(w http.ResponseWriter, r *http.Request, err interface{}) {
	ctx := r.Context()
	w.WriteHeader(http.StatusInternalServerError)
	templates.Error(UserFromFromContext(ctx).UserID != 0).Render(ctx, w)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.WriteHeader(http.StatusNotFound)
	templates.NotFound(UserFromFromContext(ctx).UserID != 0).Render(ctx, w)
}
