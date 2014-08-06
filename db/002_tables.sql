CREATE TABLE `user` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `email` varchar(100) NOT NULL UNIQUE,
  `password_hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `active` BOOLEAN NOT NULL DEFAULT 0,
  `update_token` char(36) UNIQUE,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `domain` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `user_domain` varchar(100) NOT NULL UNIQUE,
  `ip` varchar(15),
  `update_token` char(36) NOT NULL UNIQUE,
  `user_id` integer NOT NULL,
  `last_update` DATETIME NULL,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE `service` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `protocol` varchar(20) NOT NULL,
  `type` varchar(100) NOT NULL,
  `url` varchar(100),
  `port` int(11) NOT NULL,
  `domain_id` integer NOT NULL,
  `timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
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

update db_version set version = '002';
