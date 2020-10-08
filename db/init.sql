CREATE TABLE `user` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `email` varchar(100) NOT NULL UNIQUE,
  `password_hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `active` BOOLEAN NOT NULL DEFAULT 0,
  `update_token` char(36) UNIQUE,
  `unsubscribed` BOOLEAN NOT NULL DEFAULT 0,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `premium_account` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `user_id` integer NOT NULL,
  FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE `domain` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `user_domain` varchar(100) NOT NULL UNIQUE,
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
  FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE `custom_domain` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `domain` varchar(100) NOT NULL UNIQUE,
  `ip` varchar(45),
  `ipv6` varchar(45),
  `dkim_key` varchar(256),
  `update_token` char(36) UNIQUE,
  `user_id` integer NOT NULL,
  `port` integer,
  `last_update` DATETIME NULL,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
    version varchar(10) not null,
    `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

insert into db_version (version) values ('008');
