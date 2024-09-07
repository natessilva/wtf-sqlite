create table dial(
    id integer primary key autoincrement,
    user_id integer not null references user(id),
    name text not null,
    value int not null default 0,
    created_at datetime not null default current_timestamp,
    modified_at datetime not null default current_timestamp
);

-- always directly access individual values by both user_id and id
create index dial_user_id_idx on dial(user_id, id);