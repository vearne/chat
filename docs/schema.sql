
-- 2019-11-18
CREATE TABLE `account` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nickname` varchar(30) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `broker` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `modified_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

CREATE TABLE `session` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `status` int(11) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `modified_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;


CREATE TABLE `session_account` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `session_id` bigint unsigned DEFAULT NULL,
  `account_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

ALTER TABLE session_account ADD index idx_session (`session_id`);

CREATE TABLE `outbox` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `sender_id` bigint unsigned DEFAULT NULL,
  `session_id` bigint unsigned DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `msg_type` int(11) DEFAULT NULL,
  `content` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `modified_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

ALTER TABLE outbox ADD index idx_sender (`sender_id`);

CREATE TABLE `inbox` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `sender_id` bigint unsigned DEFAULT NULL,
  `msg_id` bigint unsigned DEFAULT NULL,
  `receiver_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

ALTER TABLE inbox ADD index idx_receiver (`receiver_id`);

-- 2019-12-13
CREATE TABLE `view_ack` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `session_id` bigint unsigned DEFAULT NULL,
  `account_id` bigint unsigned DEFAULT NULL,
  `msg_id` bigint unsigned DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

ALTER TABLE view_ack ADD unique idx_view_ack (`session_id`, `account_id`);