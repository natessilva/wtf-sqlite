create table user(
  id integer primary key autoincrement,
  user_name text not null unique,
  password blob not null,
  created_at datetime not null default current_timestamp
);