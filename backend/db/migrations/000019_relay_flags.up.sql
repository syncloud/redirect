alter table domain add column relay tinyint(1) not null default 0;
alter table domain add column mail_relay tinyint(1) not null default 0;
