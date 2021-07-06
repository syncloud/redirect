alter table domain add column hosted_zone_id varchar(100) NULL;
update domain set hosted_zone_id = '0';
alter table domain modify hosted_zone_id varchar(100) NOT NULL;
insert into db_version (version) values ('011');
