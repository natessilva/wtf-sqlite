create table session (
    id text primary key,
    user_id integer not null references user(id),
    created_at datetime not null default current_timestamp,
    expires_at datetime not null
);

create index session_expires_at_idx on session(expires_at);