package sqlite

import (
	"net/http"
	"sqlite/templates"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	http.Handler
	AuthService *AuthService
	UserService *UserService
}

func NewHandler(authService *AuthService, userService *UserService) *Handler {
	mux := http.NewServeMux()
	h := &Handler{
		Handler:     mux,
		AuthService: authService,
		UserService: userService,
	}

	router := httprouter.New()

	// if you are already authenticated, none of these routes make
	// sense
	router.GET("/login", requireNoAuth(h.handleGetLogin))
	router.POST("/login", requireNoAuth(h.handlePostLogin))
	router.GET("/signup", requireNoAuth(h.handleGetSignup))
	router.POST("/signup", requireNoAuth(h.handlePostSignup))

	// these routes are public.
	router.GET("/logout", h.handleLogout)
	router.GET("/", h.handleIndex)

	mux.Handle("/", authService.Middleware(router))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	return h
}

func requireNoAuth(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		userId := UserFromFromContext(r.Context())
		if userId != 0 {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		handle(w, r, p)
	}
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := UserFromFromContext(r.Context())
	if userId == 0 {
		templates.Index("World").Render(r.Context(), w)
		return
	}
	user, err := h.UserService.Get(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	templates.Index(user.UserName).Render(r.Context(), w)
}

func (h *Handler) handleGetLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	templates.Login("", "").Render(r.Context(), w)
}

func (h *Handler) handlePostLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		handleError(w, err)
		return
	}
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	output, err := h.AuthService.Login(r.Context(), AuthInput{
		UserName: userName,
		Password: password,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	if !output.OK {
		w.WriteHeader(http.StatusUnauthorized)
		templates.Login("Invalid email and/or password", userName).Render(r.Context(), w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    output.Token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) handleGetSignup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	templates.Signup("", "").Render(r.Context(), w)
}

func (h *Handler) handlePostSignup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		handleError(w, err)
		return
	}
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	if userName == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		templates.Signup("Missing required values", "").Render(r.Context(), w)
		return
	}
	output, err := h.AuthService.Signup(r.Context(), AuthInput{
		UserName: userName,
		Password: password,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	if !output.OK {
		w.WriteHeader(http.StatusUnauthorized)
		templates.Signup("Username already claimed", "").Render(r.Context(), w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    output.Token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusFound)
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

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
}
