-- name: CreateSession :exec
insert into session(id, user_id, ttl)
values(?,?,?);

-- name: DeleteSession :exec
delete from session where id = ?;

-- name: GetSession :one
select user_id, datetime(created_at, '+' || ttl || ' days') < current_timestamp as expired from session where id = ?;