create table if not exists Types (
  id integer primary key,
  type varchar(32) not null
);

create table if not exists Tasks (
  id integer primary key,
  alias varchar(64) not null,
  description varchar(128),
  type_id integer not null,
  task_timestamp integer not null,
  estimate_time varchar(24),
  real_time varchar(24),
  foreign key (type_id) references Types(id)
);

create table if not exists Tags (
  id integer primary key,
  task_id integer not null,
  tag varchar(32) not null,
  foreign key (task_id) references Tasks(id)
);

create table if not exists Reminders (
  id integer primary key,
  task_id integer not null,
  reminder varchar(24) not null,
  foreign key (task_id) references Tasks(id)
);
