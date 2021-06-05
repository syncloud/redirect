alter table domain add column domain varchar(100) NULL;
update domain set domain = concat(user_domain, ".syncloud.it");
alter table domain modify domain varchar(100) NOT NULL UNIQUE;
alter table domain change user_domain deprecated_user_domain varchar(100) NULL;
insert into db_version (version) values ('010');

