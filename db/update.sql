set @has_subscription_type := (select count(*) from information_schema.columns
  where table_schema = database() and table_name = 'user' and column_name = 'subscription_type');
set @stmt := if(@has_subscription_type = 0,
  'alter table user add column `subscription_type` integer NULL', 'do 0');
prepare add_subscription_type from @stmt;
execute add_subscription_type;
deallocate prepare add_subscription_type;
insert into db_version (version) values ('015');

create table if not exists relay_traffic (
  `name` varchar(255) not null,
  `year_month` char(7) not null,
  `bytes` bigint not null default 0,
  primary key (`name`, `year_month`)
);
insert into db_version (version) values ('016');
