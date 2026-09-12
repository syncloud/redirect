alter table `action` add column `created_at` timestamp null;
update `action` set `created_at` = coalesce(`sent_at`, `timestamp`);
alter table `action` modify column `created_at` timestamp not null default current_timestamp;
create index `action_type_created_at` on `action` (`action_type_id`, `created_at`);
