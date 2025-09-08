-- name: CreateTask :one
insert into task(user_id, title, description, ordinal) values(?,?,?,?) returning *;

-- name: ListTasks :many
select * from task where user_id = ? order by ordinal;

-- name: GetTask :one
select * from task where user_id = ? and id = ?;

-- name: UpdateTask :exec
update task set title = ?, description = ?, is_completed = ?, modified_at = current_timestamp where id = ?;

-- name: DeleteTask :exec
delete from task where id = ?;

-- name: GetMaxTaskOrdinal :one
select ordinal from task where user_id = ? order by ordinal desc limit 1;

-- name: GetPreviousTaskOrdinal :one
select ordinal from task where user_id = ? and ordinal < ? order by ordinal desc limit 1;

-- name: SetTaskOrdinal :exec
update task set ordinal = ? where id = ?;