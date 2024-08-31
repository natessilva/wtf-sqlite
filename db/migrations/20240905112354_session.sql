create table session (
    id text primary key,
    user_id integer not null references user(id),
    created_at datetime not null default current_timestamp,
    ttl integer not null
);

create index session_expires_at_idx on session(datetime(created_at, '+' || ttl || ' days'));