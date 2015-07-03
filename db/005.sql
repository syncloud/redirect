insert into `user` (`id`, `email`, `password_hash`, `active`, `update_token`, `unsubscribed`, `timestamp`)
select `id`, `email`, `password_hash`, `active`, `update_token`, 0 as `unsubscribed`, `timestamp`
from redirect_backup.`user`;

insert into `domain`
select *
from redirect_backup.`domain`;

insert into `service`
select *
from redirect_backup.`service`;

insert into `action`
select * from redirect_backup.`action`;
