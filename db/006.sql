insert into `user` select * from redirect_backup.`user`;
insert into `domain` select * from redirect_backup.`domain`;
insert into `service` select * from redirect_backup.`service`;
insert into `action` select * from redirect_backup.`action`;