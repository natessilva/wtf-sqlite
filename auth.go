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
	var tuID int64
	err = svc.db.Transaction(ctx, func(ctx context.Context, q *model.Queries) error {
		userID, err := q.CreateUser(ctx, model.CreateUserParams{
			UserName: userName,
			Password: hash,
		})
		if err != nil {
			return err
		}
		teamID, err := q.CreateTeam(ctx, userName)
		if err != nil {
			return err
		}
		tuID, err = q.CreateTeamUser(ctx, model.CreateTeamUserParams{
			TeamID: teamID,
			UserID: userID,
		})
		if err != nil {
			return err
		}
		return q.SetDefaultTeamUser(ctx, model.SetDefaultTeamUserParams{
			IsDefault: true,
			ID:        tuID,
		})
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
		ID:         token,
		TeamUserID: tuID,
		ExpiresAt:  time.Now().AddDate(0, 0, 30),
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
	teamUser, err := svc.db.Queries.GetDefaultTeamUser(ctx, user.ID)
	if err != nil {
		return AuthOutput{}, err
	}
	sessionID, err := uuid.NewV4()
	if err != nil {
		return AuthOutput{}, err
	}
	token := sessionID.String()
	svc.db.Queries.CreateSession(ctx, model.CreateSessionParams{
		ID:         token,
		TeamUserID: teamUser.ID,
		ExpiresAt:  time.Now().AddDate(0, 0, 30),
	})
	// otherwise we're in
	return AuthOutput{
		Token: token,
		OK:    true,
	}, nil
}

func (svc *AuthService) GetTeamUserFromSession(ctx context.Context, token string) (model.TeamUser, error) {
	session, err := svc.db.Queries.GetSession(ctx, token)
	if err != nil {
		return model.TeamUser{}, err
	}
	if session.Expired {
		svc.db.Queries.DeleteSession(ctx, token)
		return model.TeamUser{}, nil
	}
	return svc.db.Queries.GetTeamUser(ctx, session.TeamUserID)
}

type contextKey struct{}

var key contextKey

func ContextWithUser(ctx context.Context, i model.TeamUser) context.Context {
	return context.WithValue(ctx, &key, i)
}

func RequestWithUser(r *http.Request, i model.TeamUser) *http.Request {
	return r.WithContext(ContextWithUser(r.Context(), i))
}

func UserFromFromContext(ctx context.Context) model.TeamUser {
	value := ctx.Value(&key)
	if value != nil {
		return value.(model.TeamUser)
	}
	return model.TeamUser{}
}

func (svc *AuthService) Middleware(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("token"); err == nil {
			if tu, err := svc.GetTeamUserFromSession(r.Context(), cookie.Value); err == nil && tu.ID != 0 {
				// if we got a user, put it in the request context
				r = RequestWithUser(r, tu)

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
