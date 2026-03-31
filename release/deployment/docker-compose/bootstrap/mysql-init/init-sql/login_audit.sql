CREATE TABLE IF NOT EXISTS `login_audit`
(
    `id`         bigint(20) NOT NULL COMMENT 'Primary Key ID',
    `user_id`    bigint(20) NOT NULL DEFAULT 0 COMMENT 'User ID',
    `login_type` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'Login Type (1=password, 2=Gitee OAuth)',
    `provider`   varchar(32)          DEFAULT '' COMMENT 'OAuth Provider',
    `ip`         varchar(64)          DEFAULT '' COMMENT 'IP Address',
    `user_agent` varchar(512)        DEFAULT '' COMMENT 'User Agent',
    `success`    tinyint(1) NOT NULL DEFAULT 0 COMMENT 'Success (1=success, 0=fail)',
    `fail_reason` varchar(256)       DEFAULT '' COMMENT 'Failure Reason',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT = 'Login Audit Table';
