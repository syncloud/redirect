insert into `user` select * from redirect_backup.`user`;

insert into `domain` (`id`,`user_domain`,`ip`,`local_ip`,`map_local_address`,`update_token`,`user_id`,`device_mac_address`,`device_name`,`device_title`,`web_protocol`,`web_port`,`web_local_port`,`last_update`,`timestamp`)
select d.`id`,d.`user_domain`,d.`ip`,d.`local_ip`,d.`map_local_address`,d.`update_token`,d.`user_id`,d.`device_mac_address`,d.`device_name`,d.`device_title`,s.`protocol`,coalesce(s.`port`, 0),coalesce(s.`local_port`, 0),d.`last_update`,d.`timestamp`
from redirect_backup.`domain` as d
left join (select * from redirect_backup.`service` where `name`="server") as s on d.`id`=s.`domain_id`;

insert into `action` select * from redirect_backup.`action`;

drop table service;