alter table user add column `registered_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
alter table user add column `status` integer DEFAULT 0;
insert into db_version (version) values ('013');
