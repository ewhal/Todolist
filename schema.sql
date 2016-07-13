CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `email` VARCHAR(320),
  `password` CHAR(76),
  PRIMARY KEY (`id`)
);

CREATE TABLE `tasks` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `name` longtext,
  `title` longtext,
  `task` longtext,
  `created` DATETIME,
  `duedate` DATETIME,
  `email` VARCHAR(320),
  `completed` boolean,
  `public` boolean,
  PRIMARY KEY (`id`)
);

