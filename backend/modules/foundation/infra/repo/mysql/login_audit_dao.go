// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package mysql

import (
	"context"

	"github.com/coze-dev/coze-loop/backend/infra/db"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/infra/repo/mysql/gorm_gen/model"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/infra/repo/mysql/gorm_gen/query"
)

// ILoginAuditDAO interface for login audit operations
type ILoginAuditDAO interface {
	Create(ctx context.Context, loginAudit *model.LoginAudit, opts ...db.Option) error
}

type LoginAuditDAOImpl struct {
	db    db.Provider
	query *query.Query
}

func NewLoginAuditDAOImpl(db db.Provider) ILoginAuditDAO {
	return &LoginAuditDAOImpl{
		db:    db,
		query: query.Use(db.NewSession(context.Background())),
	}
}

func (dao *LoginAuditDAOImpl) Create(ctx context.Context, loginAudit *model.LoginAudit, opts ...db.Option) error {
	dao.query = query.Use(dao.db.NewSession(ctx, opts...))
	return dao.query.LoginAudit.WithContext(ctx).Create(loginAudit)
}
