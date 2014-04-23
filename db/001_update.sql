CREATE TABLE `user_tmp` (
  `user_domain` varchar(100) NOT NULL UNIQUE,
  `update_token` char(36) NOT NULL UNIQUE,
  `ip` varchar(15),
  `port` int(11),
  `email` varchar(100) NOT NULL UNIQUE PRIMARY KEY,
  `password_hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `active` BOOLEAN NOT NULL DEFAULT 0,
  `activate_token` char(36) UNIQUE
);

insert into user_tmp (
  `user_domain`,
  `update_token`,
  `ip`,
  `port`,
  `email`,
  `password_hash`,
  `active`,
  `activate_token`)
select
  `user_domain`,
  `update_token`,
  `ip`,
  `port`,
  `email`,
  null,
  `active`,
  null
from user;

drop table user;
RENAME TABLE user_tmp TO user;