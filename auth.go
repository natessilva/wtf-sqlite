package sqlite

import (
	"context"
	"database/sql"
	"sqlite/model"
	"strings"

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
	UserID int64
	OK     bool
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
	return AuthOutput{
		UserID: userID,
		OK:     true,
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
	// otherwise we're in
	return AuthOutput{
		UserID: user.ID,
		OK:     true,
	}, nil
}
