alter table user add column `registration_timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
insert into db_version (version) values ('013');
