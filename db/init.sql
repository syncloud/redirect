CREATE TABLE premium_status (
  `id` integer PRIMARY KEY,
  `name` varchar(100) NOT NULL
);

insert into premium_status (id, name) values (1, 'inactive');
insert into premium_status (id, name) values (2, 'pending');
insert into premium_status (id, name) values (3, 'active');

CREATE TABLE `user` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `email` varchar(100) NOT NULL UNIQUE,
  `password_hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `active` BOOLEAN NOT NULL DEFAULT 0,
  `update_token` char(36) UNIQUE,
  `notification_enabled` BOOLEAN NOT NULL DEFAULT 1,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `premium_status_id` integer NOT NULL DEFAULT 1,
  `subscription_id` varchar(100) NULL,
  `registered_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `status` integer DEFAULT 0,
  `status_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (premium_status_id) REFERENCES premium_status(id),
);

CREATE TABLE `domain` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `deprecated_user_domain` varchar(100) NULL,
  `ip` varchar(45),
  `ipv6` varchar(45),
  `dkim_key` varchar(256),
  `local_ip` varchar(45),
  `map_local_address` BOOLEAN DEFAULT 0,
  `update_token` char(36) UNIQUE,
  `user_id` integer NOT NULL,
  `device_mac_address` varchar(20),
  `device_name` varchar(100),
  `device_title` varchar(100),
  `platform_version` varchar(20),
  `web_protocol` varchar(20),
  `web_port` integer,
  `web_local_port` integer,
  `last_update` DATETIME NULL,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(100) NOT NULL UNIQUE,
  `hosted_zone_id` varchar(100) NOT NULL,
  FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE `action_type` (
  `id` integer PRIMARY KEY,
  `name` varchar(100) NOT NULL
);

CREATE TABLE `action` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `action_type_id` integer NOT NULL,
  `user_id` integer NOT NULL,
  `token` char(36) NOT NULL UNIQUE,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (action_type_id) REFERENCES action_type(id),
  FOREIGN KEY (user_id) REFERENCES `user`(id)
);

insert into action_type (id, name) values (1, 'activate');
insert into action_type (id, name) values (2, 'password');

create table db_version (
  `version` varchar(10) not null,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

insert into db_version (version) values ('014');
