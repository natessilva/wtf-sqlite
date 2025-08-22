create table session (    
    id blob primary key,
    user_id integer not null references user(id),
    created_at datetime not null default current_timestamp,
    expires_at datetime not null
);
create index session_user_id_idx on session(user_id);