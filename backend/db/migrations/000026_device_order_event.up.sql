create table `device_order_event` (
    `id` int not null auto_increment,
    `order_id` int not null,
    `status` varchar(16) not null,
    `comment` varchar(500) not null default '',
    `created_at` timestamp not null default current_timestamp,
    primary key (`id`),
    key `order_id` (`order_id`),
    constraint `device_order_event_order` foreign key (`order_id`) references `device_order` (`id`) on delete cascade
) engine=InnoDB default charset=utf8mb4;
