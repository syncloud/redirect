alter table domain add column domain varchar(100) NULL;
update domain set domain = concat(user_domain, ".syncloud.it");
alter table domain modify domain varchar(100) NOT NULL UNIQUE;

insert into db_version (version) values ('010');

