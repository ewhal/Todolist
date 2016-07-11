CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `email` VARCHAR(320),
  `password` CHAR(76),
  PRIMARY KEY (`id`)
);

CREATE TABLE `tasks` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `title` longtext,
  `task` longtext,
  `created` DATETIME,
  `duedate` DATETIME,
  `email` VARCHAR(320),
  PRIMARY KEY (`id`)
);

