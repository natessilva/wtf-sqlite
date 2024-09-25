create table team(
    id integer primary key autoincrement,
    name text not null,
    created_at datetime not null default current_timestamp
);

create table team_user(
    id integer primary key autoincrement,
    team_id integer not null references team(id), 
    user_id integer not null references user(id),
    is_default boolean not null default false,

    unique(team_id, user_id)
);

create index team_user_user_id_idx on team_user(user_id);
create unique index team_user_default_uniq_idx on team_user(user_id) where is_default;

create table session (
    id text primary key,
    team_user_id integer not null references team_user(id),
    created_at datetime not null default current_timestamp,
    expires_at datetime not null
);

create index session_expires_at_idx on session(expires_at);

