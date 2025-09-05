-- name: CreateTask :one
insert into task(user_id, title, description) values(?,?,?) returning *;

-- name: ListTasks :many
select * from task where user_id = ? order by created_at;

-- name: GetTask :one
select * from task where user_id = ? and id = ?;

-- name: UpdateTask :exec
update task set title = ?, description = ?, is_completed = ?, modified_at = current_timestamp where id = ?;

-- name: DeleteTask :exec
delete from task where id = ?;