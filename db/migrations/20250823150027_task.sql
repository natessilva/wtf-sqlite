create table task(
    id integer primary key autoincrement,
    user_id integer not null references user(id),
    title text not null,
    description text not null,
    is_completed boolean not null default 0,
    created_at datetime not null default current_timestamp,
    modified_at datetime not null default current_timestamp
);

create index task_user_id_idx on task(user_id);