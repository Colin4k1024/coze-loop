// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"
)

const TableNameLoginAudit = "login_audit"

// LoginAudit LoginAudit Table
type LoginAudit struct {
	ID        int64     `gorm:"column:id;type:bigint(20);primaryKey;comment:Primary Key ID" json:"id"`                          // Primary Key ID
	UserID    int64     `gorm:"column:user_id;type:bigint(20);not null;comment:User ID" json:"user_id"`                        // User ID
	LoginType int8      `gorm:"column:login_type;type:tinyint(1);not null;comment:Login Type (1=password, 2=Gitee OAuth)" json:"login_type"` // Login Type
	Provider  string    `gorm:"column:provider;type:varchar(32);comment:OAuth Provider" json:"provider"`                         // OAuth Provider
	IP        string    `gorm:"column:ip;type:varchar(64);comment:IP Address" json:"ip"`                                         // IP Address
	UserAgent string    `gorm:"column:user_agent;type:varchar(512);comment:User Agent" json:"user_agent"`                        // User Agent
	Success   int8      `gorm:"column:success;type:tinyint(1);not null;comment:Success (1=success, 0=fail)" json:"success"`   // Success
	FailReason string   `gorm:"column:fail_reason;type:varchar(256);comment:Failure Reason" json:"fail_reason"`                  // Failure Reason
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
}

// TableName LoginAudit's table name
func (*LoginAudit) TableName() string {
	return TableNameLoginAudit
}
