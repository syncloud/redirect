-- Schema through version 017 is created by init.sql, so there are no pending
-- migrations. Add future deltas below, bumping the version at the end.

create table if not exists mail_relay_usage (
  `name` varchar(255) not null,
  `year_month` char(7) not null,
  `messages` bigint not null default 0,
  `bounces` bigint not null default 0,
  primary key (`name`, `year_month`)
);

create table if not exists mail_relay_blocked (
  `name` varchar(255) not null,
  `reason` varchar(255) not null,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  primary key (`name`)
);

insert into db_version (version) values ('018');

alter table domain add column relay tinyint(1) not null default 0;
alter table domain add column mail_relay tinyint(1) not null default 0;

insert into db_version (version) values ('019');
