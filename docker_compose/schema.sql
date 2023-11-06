create database chat;

CREATE TABLE `account` (
   `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
   `nickname` varchar(30) DEFAULT NULL,
   `status` int(11) DEFAULT NULL,
   `broker` varchar(255) DEFAULT NULL,
   `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `modified_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   `token` varchar(50) DEFAULT '',
   PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `inbox` (
     `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
     `sender_id` bigint(20) unsigned DEFAULT NULL,
     `msg_id` bigint(20) unsigned DEFAULT NULL,
     `receiver_id` bigint(20) unsigned DEFAULT NULL,
     PRIMARY KEY (`id`),
     KEY `idx_receiver` (`receiver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `outbox` (
      `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
      `sender_id` bigint(20) unsigned DEFAULT NULL,
      `session_id` bigint(20) unsigned DEFAULT NULL,
      `status` int(11) DEFAULT NULL,
      `msg_type` int(11) DEFAULT NULL,
      `content` varchar(255) DEFAULT NULL,
      `created_at` timestamp NULL DEFAULT NULL,
      `modified_at` timestamp NULL DEFAULT NULL,
      PRIMARY KEY (`id`),
      KEY `idx_sender` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `session` (
       `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
       `status` int(11) DEFAULT NULL,
       `created_at` timestamp NULL DEFAULT NULL,
       `modified_at` timestamp NULL DEFAULT NULL,
       PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `session_account` (
       `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
       `session_id` bigint(20) unsigned DEFAULT NULL,
       `account_id` bigint(20) unsigned DEFAULT NULL,
       PRIMARY KEY (`id`),
       KEY `idx_session` (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `view_ack` (
        `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
        `session_id` bigint(20) unsigned DEFAULT NULL,
        `account_id` bigint(20) unsigned DEFAULT NULL,
        `msg_id` bigint(20) unsigned DEFAULT NULL,
        `created_at` timestamp NULL DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `idx_view_ack` (`session_id`,`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;