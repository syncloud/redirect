alter table `user` add column `admin` tinyint(1) not null default 0;
alter table `device_order` add column `status` varchar(16) not null default 'ordered';
