create database redirect;
use redirect;
CREATE TABLE `user` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `email` varchar(100) NOT NULL UNIQUE,
  `password_hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `active` BOOLEAN NOT NULL DEFAULT 0,
  `activate_token` char(36) UNIQUE
);

CREATE TABLE `domain` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `user_domain` varchar(100) NOT NULL UNIQUE,
  `ip` varchar(15),
  `update_token` char(36) NOT NULL UNIQUE,
  `user_id` integer NOT NULL
);

CREATE TABLE `service` (
  `id` integer PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `type` varchar(100) NOT NULL,
  `url` varchar(100),
  `port` int(11) NOT NULL,
  `domain_id` integer NOT NULL
);