create database redirect;
use redirect;
CREATE TABLE `user` (
  `user_domain` varchar(100) NOT NULL PRIMARY KEY,
  `email` varchar(100) NOT NULL UNIQUE,
  `password_hash` varchar(44) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `update_token` char(36) NOT NULL UNIQUE,
  `ip` varchar(15) NOT NULL,
  `port` int(11) NOT NULL,
  `active` BIT(1) NOT NULL DEFAULT 0
);