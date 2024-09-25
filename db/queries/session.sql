-- name: CreateSession :exec
insert into session(id, team_user_id, expires_at)
values(?,?,?);

-- name: DeleteSession :exec
delete from session where id = ?;

-- name: GetSession :one
select team_user_id, expires_at < current_timestamp as expired from session where id = ?;