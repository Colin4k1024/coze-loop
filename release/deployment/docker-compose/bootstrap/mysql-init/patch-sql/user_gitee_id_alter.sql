-- Add gitee_id column to user table for Gitee OAuth support
ALTER TABLE `user` ADD COLUMN `gitee_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'Gitee OAuth ID' AFTER `session_key`;

-- Add unique index on gitee_id
ALTER TABLE `user` ADD UNIQUE KEY `idx_gitee_id` (`gitee_id`);
