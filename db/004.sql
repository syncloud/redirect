insert into `user`
select * from redirect_backup.`user`;

insert into `domain` (`id`, `user_domain`, `ip`, `local_ip`, `update_token`, `user_id`, `device_mac_address`, `device_name`, `device_title`, `last_update`)
select `id`, `user_domain`, `ip`, '0.0.0.0' as `local_ip`, `update_token`, `user_id`, '00:00:00:00:00:00' as `device_mac_address`, 'Not assigned' as `device_name`, 'Not assigned' as `device_title`, `last_update`
from redirect_backup.`domain`;

insert into `action`
select * from redirect_backup.`action`;
