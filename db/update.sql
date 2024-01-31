alter table user add column `subscription_type` integer NULL;
insert into db_version (version) values ('015');
