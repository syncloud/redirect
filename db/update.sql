alter table user add column `status_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
insert into db_version (version) values ('014');
