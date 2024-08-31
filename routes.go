package sqlite

import (
	"net/http"
	"sqlite/templates"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	http.Handler
	AuthService *AuthService
}

func NewHandler(authService *AuthService) *Handler {
	mux := http.NewServeMux()
	h := &Handler{
		Handler:     mux,
		AuthService: authService,
	}

	router := httprouter.New()

	// explicitly require that these routes
	// do not have an authenticated user
	// using a middleware called noAuth
	// it will redirect to "/" if there is
	// an authenticated user already.
	router.GET("/login", h.handleGetLogin)
	router.POST("/login", h.handlePostLogin)
	router.GET("/signup", h.handleGetSignup)
	router.POST("/signup", h.handlePostSignup)
	router.GET("/logout", h.handleLogout)

	// require every other route to be authenticated
	// via a middleware called auth. it will 401
	// and redirect to login
	router.GET("/", h.handleIndex)

	// assume I already have the middlewares implemented

	mux.Handle("/", router)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	return h
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	cookie, err := r.Cookie("userID")
	if err != nil {
		templates.Index("World").Render(r.Context(), w)
		return
	}
	// userID, err := strconv.ParseInt(cookie.Value, 10, 64)
	if err != nil {
		handleError(w, err)
		return
	}
	templates.Index(cookie.Value).Render(r.Context(), w)
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
		Name:     "userID",
		Value:    strconv.FormatInt(output.UserID, 10),
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
		Name:     "userID",
		Value:    strconv.FormatInt(output.UserID, 10),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	cookie, err := r.Cookie("userID")
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
