alter table user add column subscription_id varchar(100) NULL;
alter table user change unsubscribed notification_enabled BOOLEAN NOT NULL DEFAULT 0;
update user set notification_enabled = not notification_enabled;
insert into db_version (version) values ('012');
