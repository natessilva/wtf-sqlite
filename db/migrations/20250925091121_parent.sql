alter table task add column parent_id integer references task(id) on delete set null;
create index task_parent_id_ordinal_idx on task(parent_id, ordinal);
drop index task_user_id_ordinal_idx;
create index task_user_id_ordinal_idx on task(user_id, ordinal) where parent_id is null;