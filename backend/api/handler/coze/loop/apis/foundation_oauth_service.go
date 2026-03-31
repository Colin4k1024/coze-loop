// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/coze-dev/coze-loop/backend/infra/middleware/session"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/application"
	"github.com/coze-dev/coze-loop/backend/pkg/hertzutil"
)

// OAuthApplicationInterface defines the interface for OAuth operations
type OAuthApplicationInterface interface {
	GetGiteeConfig(ctx context.Context) *application.OAuthConfig
	LoginByGitee(ctx context.Context, code, state string) (int64, string, error)
}

// FoundationOAuthService handles OAuth endpoints
type FoundationOAuthService struct {
	oauthApp OAuthApplicationInterface
}

func NewFoundationOAuthService(oauthApp *application.OAuthApplication) *FoundationOAuthService {
	return &FoundationOAuthService{oauthApp: oauthApp}
}

// GiteeAuthorize handles GET /api/foundation/v1/oauth/gitee/authorize
// @router /api/foundation/v1/oauth/gitee/authorize [GET]
func (s *FoundationOAuthService) GiteeAuthorize(ctx context.Context, c *app.RequestContext) {
	config := s.oauthApp.GetGiteeConfig(ctx)
	if config == nil || !config.Enabled {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "Gitee OAuth is not enabled",
		})
		return
	}

	redirectURI := config.RedirectURI
	authorizeURL, err := generateGiteeAuthorizeURL(redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("failed to generate authorize URL: %v", err),
		})
		return
	}

	c.Redirect(http.StatusFound, authorizeURL)
}

// GiteeCallback handles GET /api/foundation/v1/oauth/gitee/callback
// @router /api/foundation/v1/oauth/gitee/callback [GET]
func (s *FoundationOAuthService) GiteeCallback(ctx context.Context, c *app.RequestContext) {
	code := string(c.Query("code"))
	state := string(c.Query("state"))

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "code and state are required",
		})
		return
	}

	userID, sessionKey, err := s.oauthApp.LoginByGitee(ctx, code, state)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": fmt.Sprintf("login failed: %v", err),
		})
		return
	}

	// Set session cookie
	c.SetCookie(session.SessionKey,
		sessionKey,
		int(session.SessionExpires),
		"/",
		hertzutil.GetOriginHost(c),
		protocol.CookieSameSiteDefaultMode,
		false,
		true)

	// Redirect to frontend with user_id as query param
	frontendURL := fmt.Sprintf("/#/oauth/callback?user_id=%d", userID)
	c.Redirect(http.StatusFound, frontendURL)
}

func generateGiteeAuthorizeURL(redirectURI string) (string, error) {
	// The actual URL generation is handled by the OAuth service
	// Here we just return a placeholder that will be replaced
	return fmt.Sprintf("https://gitee.com/oauth/authorize?redirect_uri=%s", redirectURI), nil
}

// OAuthHandler is the global OAuth handler instance
var OAuthHandler *FoundationOAuthService

// InitOAuthHandler initializes the OAuth handler
func InitOAuthHandler(oauthApp *application.OAuthApplication) {
	OAuthHandler = NewFoundationOAuthService(oauthApp)
}
