create database redirect;
use redirect;
CREATE TABLE `user` (
  `user_domain` varchar(100) NOT NULL UNIQUE,
  `update_token` char(36) NOT NULL UNIQUE,
  `ip` varchar(15),
  `port` int(11),
  `email` varchar(100) NOT NULL UNIQUE PRIMARY KEY,
  `password_hash` varchar(44) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `active` BOOLEAN NOT NULL DEFAULT 0,
  `activate_token` char(36) UNIQUE
);