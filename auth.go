package sqlite

import (
	"context"
	"database/sql"
	"net/http"
	"sqlite/model"
	"strings"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *DB
}

func NewAuthService(db *DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

type AuthInput struct {
	UserName string
	Password string
}

type AuthOutput struct {
	Token string
	OK    bool
}

func (svc *AuthService) Signup(ctx context.Context, input AuthInput) (AuthOutput, error) {
	userName := strings.ToLower(input.UserName)
	_, err := svc.db.Queries.GetUserByUsername(ctx, userName)
	if err == nil {
		return AuthOutput{
			OK: false,
		}, nil
	}
	if err != sql.ErrNoRows {
		return AuthOutput{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return AuthOutput{}, err
	}
	userID, err := svc.db.Queries.CreateUser(ctx, model.CreateUserParams{
		UserName: userName,
		Password: hash,
	})
	if err != nil {
		return AuthOutput{}, err
	}
	sessionID, err := uuid.NewV4()
	if err != nil {
		return AuthOutput{}, err
	}
	token := sessionID.String()
	svc.db.Queries.CreateSession(ctx, model.CreateSessionParams{
		ID:     token,
		UserID: userID,
		Ttl:    30,
	})
	return AuthOutput{
		Token: token,
		OK:    true,
	}, nil
}

func (svc *AuthService) Login(ctx context.Context, input AuthInput) (AuthOutput, error) {
	userName := strings.ToLower(input.UserName)
	user, err := svc.db.Queries.GetUserByUsername(ctx, userName)
	if err != nil {
		// if the user doesn't exist they cannot login
		if err == sql.ErrNoRows {
			return AuthOutput{OK: false}, nil
		}
		// otherwise something unexpected happened
		return AuthOutput{}, err
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(input.Password))
	if err != nil {
		// if the password doesn't match they cannot login
		return AuthOutput{OK: false}, nil
	}
	sessionID, err := uuid.NewV4()
	if err != nil {
		return AuthOutput{}, err
	}
	token := sessionID.String()
	svc.db.Queries.CreateSession(ctx, model.CreateSessionParams{
		ID:     token,
		UserID: user.ID,
		Ttl:    30,
	})
	// otherwise we're in
	return AuthOutput{
		Token: token,
		OK:    true,
	}, nil
}

type Session struct {
	UserID  int64
	Expired bool
}

func (svc *AuthService) GetUserFromSession(ctx context.Context, token string) (int64, error) {
	session, err := svc.db.Queries.GetSession(ctx, token)
	if err != nil {
		return 0, err
	}
	return session.UserID, nil
}

type contextKey struct{}

var key contextKey

type User struct {
	ID int64
}

func RequestWithUser(r *http.Request, i int64) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), &key, i))
}

func UserFromFromContext(ctx context.Context) int64 {
	value := ctx.Value(&key)
	if value != nil {
		return value.(int64)
	}
	return 0
}

func (svc *AuthService) Middleware(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("token"); err == nil {
			if userId, err := svc.GetUserFromSession(r.Context(), cookie.Value); err == nil {
				r = RequestWithUser(r, userId)
			}
		}
		handle.ServeHTTP(w, r)
	})
}
