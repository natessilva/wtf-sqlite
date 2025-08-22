-- name: CreateSession :exec
insert into session(id, user_id, expires_at) values(?,?,?);

-- name: DeleteSession :exec
delete from session where id = ?;

-- name: GetSession :one
select user_id, expires_at < current_timestamp as expired from session where id = ?;

-- name: DeleteExpiredSessions :exec
delete from session where id in(select id from session where expires_at < current_timestamp limit 100);

-- name: DeletedExpiredSessionsCount :one
select changes() as deleted_count;