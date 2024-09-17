-- name: CreateSession :exec
insert into session(id, user_id, expires_at)
values(?,?,?);

-- name: DeleteSession :exec
delete from session where id = ?;

-- name: GetSession :one
select user_id, expires_at < current_timestamp as expired from session where id = ?;