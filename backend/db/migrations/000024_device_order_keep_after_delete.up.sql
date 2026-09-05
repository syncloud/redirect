alter table `device_order` drop foreign key `device_order_user`;
alter table `device_order` modify column `user_id` int null;
alter table `device_order` add constraint `device_order_user` foreign key (`user_id`) references `user` (`id`) on delete set null;
