-- name: CreateUser :one
insert into user(user_name, password)
values(?,?)
returning id;

-- name: GetUserByUsername :one
select * from user where user_name = ?;

-- name: GetUserById :one
select * from user where id = ?;

-- name: UpdateUser :exec
update user set user_name = ? where id = ?;

-- name: SetPassword :exec
update user set password = ? where id = ?;