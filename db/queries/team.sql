-- name: CreateTeam :one
insert into team(name) values(?) returning id;

-- name: CreateTeamUser :one
insert into team_user(team_id,user_id) values(?,?) returning id;

-- name: SetDefaultTeamUser :exec
update team_user set is_default = ? where id = ?;

-- name: GetDefaultTeamUser :one
select * from team_user where user_id = ? and is_default;

-- name: GetTeamUser :one
select * from team_user where id = ?;

-- name: ListTeams :many
select team_user.id as team_user_id, sqlc.embed(team)
from team_user
join team on team_user.team_id = team.id
where team_user.user_id = ?