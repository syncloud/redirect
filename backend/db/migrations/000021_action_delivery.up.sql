alter table `action` add column `sent_at` timestamp null;
alter table `action` add column `attempts` integer not null default 0;
update `action` set `sent_at` = `timestamp` where `sent_at` is null;
