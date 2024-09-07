-- name: CreateDial :one
insert into dial(user_id, name)
values(?,?)
returning id;

-- name: ListDials :many
select * from dial
where user_id = ?
order by modified_at desc;

-- name: GetDial :one
select * from dial where user_id = ? and id = ?;

-- name: UpdateDial :exec
update dial set name = ? where id = ?
returning id;

-- name: SetDialValue :exec
update dial set value = ? where id = ?
returning id;


-- name: DeleteDial :exec
delete from dial where id = ?
returning id;