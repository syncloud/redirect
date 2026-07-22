create table relay_traffic (
  `name` varchar(255) not null,
  `year_month` char(7) not null,
  `bytes` bigint not null default 0,
  primary key (`name`, `year_month`)
);
insert into db_version (version) values ('016');
