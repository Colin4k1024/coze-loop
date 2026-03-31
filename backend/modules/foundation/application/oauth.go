// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/coze-dev/coze-loop/backend/modules/foundation/domain/user/service"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/infra/repo/mysql"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/infra/repo/mysql/gorm_gen/model"
	"github.com/coze-dev/coze-loop/backend/pkg/conf"
	"github.com/coze-dev/coze-loop/backend/pkg/logs"
)

// OAuthConfig represents the OAuth configuration
type OAuthConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURI  string   `mapstructure:"redirect_uri"`
	Scopes       []string `mapstructure:"scopes"`
}

type oauthController struct {
	configLoader conf.IConfigLoader
}

func newOAuthController(configFactory conf.IConfigLoaderFactory) *oauthController {
	ctrl := &oauthController{}
	if loader, err := configFactory.NewConfigLoader("foundation.yaml"); err == nil {
		ctrl.configLoader = loader
	}
	return ctrl
}

func (c *oauthController) getGiteeConfig(ctx context.Context) *OAuthConfig {
	if c.configLoader == nil {
		logs.CtxWarn(ctx, "oauthController configLoader is nil")
		return nil
	}

	const keyOAuthGitee = "oauth.gitee"
	var config OAuthConfig
	if err := c.configLoader.UnmarshalKey(ctx, keyOAuthGitee, &config); err != nil {
		logs.CtxWarn(ctx, "load oauth.gitee config fail, err: %v", err)
		return nil
	}
	return &config
}

// OAuthApplication handles OAuth login operations
type OAuthApplication struct {
	giteeOAuthService service.IGiteeOAuthService
	loginAuditDAO     mysql.ILoginAuditDAO
	oauthCtrl         *oauthController
}

func NewOAuthApplication(
	giteeOAuthService service.IGiteeOAuthService,
	loginAuditDAO mysql.ILoginAuditDAO,
	configFactory conf.IConfigLoaderFactory,
) *OAuthApplication {
	return &OAuthApplication{
		giteeOAuthService: giteeOAuthService,
		loginAuditDAO:     loginAuditDAO,
		oauthCtrl:         newOAuthController(configFactory),
	}
}

// GetGiteeConfig returns the Gitee OAuth configuration
func (o *OAuthApplication) GetGiteeConfig(ctx context.Context) *service.OAuthGiteeConfig {
	config := o.oauthCtrl.getGiteeConfig(ctx)
	if config == nil {
		return nil
	}
	return &service.OAuthGiteeConfig{
		Enabled:      config.Enabled,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURI:  config.RedirectURI,
		Scopes:       config.Scopes,
	}
}

// LoginByGitee handles Gitee OAuth login callback
func (o *OAuthApplication) LoginByGitee(ctx context.Context, code, state string) (userID int64, sessionKey string, err error) {
	// Validate state and exchange code for access token
	giteeUser, err := o.giteeOAuthService.HandleCallback(ctx, code, state)
	if err != nil {
		// Record failed login - need to convert to model.LoginAudit
		o.loginAuditDAO.Create(ctx, &model.LoginAudit{
			UserID:     0,
			LoginType:  2,
			Provider:   "gitee",
			Success:    0,
			FailReason: err.Error(),
		})
		return 0, "", err
	}

	// Get or create user
	userID, sessionKey, err = o.giteeOAuthService.GetOrCreateUser(ctx, giteeUser)
	if err != nil {
		// Record failed login
		o.loginAuditDAO.Create(ctx, &model.LoginAudit{
			UserID:     userID,
			LoginType:  2,
			Provider:   "gitee",
			Success:    0,
			FailReason: err.Error(),
		})
		return 0, "", err
	}

	// Record successful login
	o.loginAuditDAO.Create(ctx, &model.LoginAudit{
		UserID:    userID,
		LoginType: 2,
		Provider:  "gitee",
		Success:   1,
	})

	return userID, sessionKey, nil
}
