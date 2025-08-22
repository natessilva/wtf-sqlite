-- name: CreateUser :one
insert into user(user_name, password) values(?,?) returning id;

-- name: GetUserByUsername :one
select user.* from user where user_name = ?;

-- name: GetUserById :one
select * from user where id = ?;