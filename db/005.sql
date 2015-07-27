insert into `user` (`id`, `email`, `password_hash`, `active`, `update_token`, `unsubscribed`, `timestamp`)
select `id`, `email`, `password_hash`, `active`, `update_token`, 0 as `unsubscribed`, `timestamp`
from redirect_backup.`user`;

insert into `domain`
(`id`,
`user_domain`,
`ip`,
`local_ip`,
`map_local_address`,
`update_token`,
`user_id`,
`device_mac_address`,
`device_name`,
`device_title`,
`last_update`,
`timestamp`)
select
`id`,
`user_domain`,
`ip`,
`local_ip`,
0 as `map_local_address`,
`update_token`,
`user_id`,
`device_mac_address`,
`device_name`,
`device_title`,
`last_update`,
`timestamp`
from redirect_backup.`domain`;

insert into `service`
select *
from redirect_backup.`service`;

insert into `action`
select * from redirect_backup.`action`;
