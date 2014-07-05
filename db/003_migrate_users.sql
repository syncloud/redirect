insert into user (`email`, `password_hash`, `active`)
select `email`, `password_hash`, `active`
from redirect_backup.user;

insert into domain (user_domain, ip, update_token, user_id)
select bu.user_domain, bu.ip, bu.update_token, u.id from redirect_backup.user bu inner join user u on bu.email = u.email;

-- url is not used
insert into service (name, type, port, domain_id)
select 'owncloud', '_http._tcp', bu.port, d.id
from redirect_backup.user bu
inner join user u on bu.email = u.email
inner join domain d on u.id = d.user_id;

insert into action_type (id, name) values (1, 'activate');
insert into action_type (id, name) values (2, 'password');

update db_version set version = '003';