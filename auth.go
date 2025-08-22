package sqlite

import (
	"context"
	"database/sql"
	"net/http"
	"sqlite/model"
	"strings"
	"time"

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
	userId, err := svc.db.Queries.CreateUser(ctx, model.CreateUserParams{
		UserName: userName,
		Password: hash,
	})
	if err != nil {
		return AuthOutput{}, err
	}
	sessionID, err := uuid.NewV7()
	if err != nil {
		return AuthOutput{}, err
	}
	token := sessionID.Bytes()
	svc.db.Queries.CreateSession(ctx, model.CreateSessionParams{
		ID:        token,
		UserID:    userId,
		ExpiresAt: time.Now().AddDate(0, 0, 30),
	})
	return AuthOutput{
		Token: sessionID.String(),
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
	sessionID, err := uuid.NewV7()
	if err != nil {
		return AuthOutput{}, err
	}
	token := sessionID.Bytes()
	svc.db.Queries.CreateSession(ctx, model.CreateSessionParams{
		ID:        token,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 30),
	})
	// otherwise we're in
	return AuthOutput{
		Token: sessionID.String(),
		OK:    true,
	}, nil
}

func (svc *AuthService) GetUserFromSession(ctx context.Context, tokenString string) (model.User, error) {
	token, err := uuid.FromString(tokenString)
	if err != nil {
		return model.User{}, err
	}

	session, err := svc.db.Queries.GetSession(ctx, token.Bytes())
	if err != nil {
		return model.User{}, err
	}
	if session.Expired {
		svc.db.Queries.DeleteSession(ctx, token.Bytes())
		return model.User{}, nil
	}
	return svc.db.Queries.GetUserById(ctx, session.UserID)
}

type contextKey struct{}

var key contextKey

func ContextWithUser(ctx context.Context, i model.User) context.Context {
	return context.WithValue(ctx, &key, i)
}

func RequestWithUser(r *http.Request, i model.User) *http.Request {
	return r.WithContext(ContextWithUser(r.Context(), i))
}

func UserFromContext(ctx context.Context) model.User {
	value := ctx.Value(&key)
	if value != nil {
		return value.(model.User)
	}
	return model.User{}
}

func (svc *AuthService) Middleware(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("token"); err == nil {
			if user, err := svc.GetUserFromSession(r.Context(), cookie.Value); err == nil && user.ID != 0 {
				// if we got a user, put it in the request context
				r = RequestWithUser(r, user)

			} else if err != sql.ErrNoRows {
				// ErrNoRows just means that there isn't a session
				// any other error means something unexpected happened
				handleError(w, r, err)
				return
			}
		}
		handle.ServeHTTP(w, r)
	})
}
