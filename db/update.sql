alter table domain add column name varchar(100) NULL;
update domain set name = concat(user_domain, ".syncloud.it");
alter table domain modify name varchar(100) NOT NULL UNIQUE;
alter table domain change user_domain deprecated_user_domain varchar(100) NULL;
insert into db_version (version) values ('010');

