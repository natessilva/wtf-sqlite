drop table session;
drop table team_user;
drop table team;

create table session (    
    id blob primary key,
    user_id integer not null references user(id),
    created_at datetime not null default current_timestamp,
    expires_at datetime not null
);