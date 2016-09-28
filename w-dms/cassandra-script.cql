create table if not exists tasks.Types (
  id uuid primary key,
  type varchar
);

create index if not exists on tasks.Types (type);

create table if not exists tasks.Tasks (
  id uuid primary key,
  alias text,
  description text,
  type_id uuid,
  task_timestamp int,
  estimate_time varchar,
  real_time varchar,
  reminders set<varchar>
);

create index if not exists on tasks.Tasks (alias);

create table if not exists tasks.Tags (
  id uuid primary key,
  task_id uuid,
  tag varchar
);

create index if not exists on tasks.Tags (task_id);
