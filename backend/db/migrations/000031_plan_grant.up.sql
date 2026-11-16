create table `plan_grant` (
    `id` int not null auto_increment,
    `code` varchar(32) not null,
    `plan` varchar(16) not null,
    `months` int not null,
    `uses_left` int not null,
    `source` varchar(64) not null,
    `created_at` timestamp not null default current_timestamp,
    primary key (`id`),
    unique key `plan_grant_code` (`code`)
) engine=InnoDB default charset=utf8mb4;

create table `plan_grant_redemption` (
    `id` int not null auto_increment,
    `grant_id` int not null,
    `user_id` int null,
    `plan` varchar(16) not null,
    `ends_at` datetime not null,
    `created_at` timestamp not null default current_timestamp,
    primary key (`id`),
    key `grant_id` (`grant_id`),
    key `user_id` (`user_id`),
    constraint `plan_grant_redemption_grant` foreign key (`grant_id`) references `plan_grant` (`id`) on delete cascade,
    constraint `plan_grant_redemption_user` foreign key (`user_id`) references `user` (`id`) on delete set null
) engine=InnoDB default charset=utf8mb4;
