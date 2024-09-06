// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: session.sql

package model

import (
	"context"
)

const createSession = `-- name: CreateSession :exec
insert into session(id, user_id, ttl)
values(?,?,?)
`

type CreateSessionParams struct {
	ID     string
	UserID int64
	Ttl    int64
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.ExecContext(ctx, createSession, arg.ID, arg.UserID, arg.Ttl)
	return err
}

const deleteSession = `-- name: DeleteSession :exec
delete from session where id = ?
`

func (q *Queries) DeleteSession(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteSession, id)
	return err
}

const getSession = `-- name: GetSession :one
select user_id, datetime(created_at, '+' || ttl || ' days') < current_timestamp as expired from session where id = ?
`

type GetSessionRow struct {
	UserID  int64
	Expired bool
}

func (q *Queries) GetSession(ctx context.Context, id string) (GetSessionRow, error) {
	row := q.db.QueryRowContext(ctx, getSession, id)
	var i GetSessionRow
	err := row.Scan(&i.UserID, &i.Expired)
	return i, err
}
