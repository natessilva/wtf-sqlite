// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user.sql

package model

import (
	"context"
)

const createUser = `-- name: CreateUser :one
insert into user(user_name, password)
values(?,?)
returning id
`

type CreateUserParams struct {
	UserName string
	Password []byte
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.UserName, arg.Password)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const getUserById = `-- name: GetUserById :one
select id, user_name, password, created_at from user where id = ?
`

func (q *Queries) GetUserById(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.UserName,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
select id, user_name, password, created_at from user where user_name = ?
`

func (q *Queries) GetUserByUsername(ctx context.Context, userName string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, userName)
	var i User
	err := row.Scan(
		&i.ID,
		&i.UserName,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const setPassword = `-- name: SetPassword :exec
update user
set password = ?
where id = ?
`

type SetPasswordParams struct {
	Password []byte
	ID       int64
}

func (q *Queries) SetPassword(ctx context.Context, arg SetPasswordParams) error {
	_, err := q.db.ExecContext(ctx, setPassword, arg.Password, arg.ID)
	return err
}

const updateUser = `-- name: UpdateUser :exec
update user
set user_name = ?
where id = ?
`

type UpdateUserParams struct {
	UserName string
	ID       int64
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.ExecContext(ctx, updateUser, arg.UserName, arg.ID)
	return err
}