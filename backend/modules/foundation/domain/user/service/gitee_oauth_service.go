// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coze-dev/coze-loop/backend/infra/idgen"
	"github.com/coze-dev/coze-loop/backend/infra/middleware/session"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/domain/user/entity"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/domain/user/repo"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/pkg/errno"
	"github.com/coze-dev/coze-loop/backend/modules/foundation/pkg/pswd"
	"github.com/coze-dev/coze-loop/backend/pkg/errorx"
	"github.com/coze-dev/coze-loop/backend/pkg/json2"
	"github.com/coze-dev/coze-loop/backend/pkg/logs"
)

const (
	giteeAuthorizeURL = "https://gitee.com/oauth/authorize"
	giteeTokenURL     = "https://gitee.com/oauth/token"
	giteeUserInfoURL  = "https://gitee.com/api/v5/user"

	loginTypeGitee = 2

	stateTimeout = 10 * time.Minute
)

type IGiteeOAuthService interface {
	GenerateAuthorizeURL(ctx context.Context, redirectURI string) (string, error)
	HandleCallback(ctx context.Context, code, state string) (*GiteeUserInfo, error)
	GetOrCreateUser(ctx context.Context, giteeUser *GiteeUserInfo) (int64, string, error)
}

// OAuthConfigProvider interface for getting OAuth config at runtime
type OAuthConfigProvider interface {
	GetGiteeConfig(ctx context.Context) *OAuthGiteeConfig
}

// OAuthGiteeConfig represents Gitee OAuth configuration
type OAuthGiteeConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

type GiteeOAuthServiceImpl struct {
	userRepo     repo.IUserRepo
	idgen        idgen.IIDGenerator
	configProvider OAuthConfigProvider
	stateSecret  []byte
}

func NewGiteeOAuthService(
	userRepo repo.IUserRepo,
	idgen idgen.IIDGenerator,
	configProvider OAuthConfigProvider,
) IGiteeOAuthService {
	return &GiteeOAuthServiceImpl{
		userRepo:     userRepo,
		idgen:        idgen,
		configProvider: configProvider,
		stateSecret:  []byte("gitee-oauth-state-hmac-key"),
	}
}

func (s *GiteeOAuthServiceImpl) getConfig(ctx context.Context) (*OAuthGiteeConfig, error) {
	config := s.configProvider.GetGiteeConfig(ctx)
	if config == nil {
		return nil, fmt.Errorf("gitee oauth config is nil")
	}
	if !config.Enabled {
		return nil, fmt.Errorf("gitee oauth is not enabled")
	}
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("gitee oauth client_id or client_secret is empty")
	}
	return config, nil
}

// GiteeUserInfo represents the user info from Gitee
type GiteeUserInfo struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GenerateAuthorizeURL generates the Gitee OAuth authorize URL with state parameter
func (s *GiteeOAuthServiceImpl) GenerateAuthorizeURL(ctx context.Context, redirectURI string) (string, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("get gitee config failed"))
	}

	// Generate and sign state
	state := &OAuthState{
		RedirectURI: redirectURI,
		ExpiresAt:   time.Now().Add(stateTimeout).Unix(),
		Random:      fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	stateData, err := json.Marshal(state)
	if err != nil {
		return "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("marshal state failed"))
	}

	// Sign the state
	h := hmac.New(sha256.New, s.stateSecret)
	h.Write(stateData)
	signature := h.Sum(nil)

	// Combine state data and signature
	finalState := base64.RawURLEncoding.EncodeToString(append(stateData, signature...))

	// Build authorize URL
	authorizeURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=user_info",
		giteeAuthorizeURL,
		config.ClientID,
		url.QueryEscape(redirectURI),
	)

	logs.CtxInfo(ctx, "generate gitee authorize url: %s", authorizeURL)
	return authorizeURL + "&state=" + finalState, nil
}

// OAuthState represents the signed state parameter
type OAuthState struct {
	RedirectURI string `json:"redirect_uri"`
	ExpiresAt   int64  `json:"expires_at"`
	Random      string `json:"random"`
}

// HandleCallback handles the OAuth callback from Gitee
func (s *GiteeOAuthServiceImpl) HandleCallback(ctx context.Context, code, state string) (*GiteeUserInfo, error) {
	// Validate and parse state
	stateData, err := s.validateState(state)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInvalidParamCode, errorx.WithExtraMsg("invalid state"))
	}

	// Exchange code for access token
	tokenResp, err := s.exchangeToken(ctx, code, stateData.RedirectURI)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("exchange token failed"))
	}

	// Get user info from Gitee
	giteeUser, err := s.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("get user info failed"))
	}

	logs.CtxInfo(ctx, "gitee user info: id=%d, login=%s, name=%s, email=%s",
		giteeUser.ID, giteeUser.Login, giteeUser.Name, giteeUser.Email)

	return giteeUser, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	CreatedAt   int64  `json:"created_at"`
}

func (s *GiteeOAuthServiceImpl) validateState(state string) (*OAuthState, error) {
	// Decode state
	data, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return nil, errorx.New("decode state failed")
	}

	if len(data) < 32 {
		return nil, errorx.New("state data too short")
	}

	// Split data and signature
	stateData := data[:len(data)-32]
	signature := data[len(data)-32:]

	// Verify signature
	h := hmac.New(sha256.New, s.stateSecret)
	h.Write(stateData)
	expectedSig := h.Sum(nil)

	if !hmac.Equal(signature, expectedSig) {
		return nil, errorx.New("invalid state signature")
	}

	// Parse state
	var oauthState OAuthState
	if err := json.Unmarshal(stateData, &oauthState); err != nil {
		return nil, errorx.New("unmarshal state failed")
	}

	// Check expiration
	if time.Now().Unix() > oauthState.ExpiresAt {
		return nil, errorx.New("state expired")
	}

	return &oauthState, nil
}

func (s *GiteeOAuthServiceImpl) exchangeToken(ctx context.Context, code, redirectURI string) (*tokenResponse, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("get gitee config failed"))
	}

	// Build form data
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", redirectURI)
	formData.Set("client_id", config.ClientID)
	formData.Set("client_secret", config.ClientSecret)

	// Send POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, giteeTokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("create token request failed"))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("exchange token request failed"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errorx.NewByCode(errno.CommonInternalErrorCode, errorx.WithExtraMsg(fmt.Sprintf("token response status: %d", resp.StatusCode)))
	}

	var tokenResp tokenResponse
	if err := json2.UnmarshalFromReader(resp.Body, &tokenResp); err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("unmarshal token response failed"))
	}

	logs.CtxDebug(ctx, "gitee token response: access_token=%s, token_type=%s, expires_in=%d",
		tokenResp.AccessToken, tokenResp.TokenType, tokenResp.ExpiresIn)

	return &tokenResp, nil
}

func (s *GiteeOAuthServiceImpl) getUserInfo(ctx context.Context, accessToken string) (*GiteeUserInfo, error) {
	// Create request with Authorization header
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, giteeUserInfoURL, nil)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("create user info request failed"))
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("get user info request failed"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errorx.NewByCode(errno.CommonInternalErrorCode, errorx.WithExtraMsg(fmt.Sprintf("user info response status: %d", resp.StatusCode)))
	}

	var giteeUser GiteeUserInfo
	if err := json2.UnmarshalFromReader(resp.Body, &giteeUser); err != nil {
		return nil, errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("unmarshal user info failed"))
	}

	return &giteeUser, nil
}

// GetOrCreateUser gets existing user by gitee_id or creates a new user
func (s *GiteeOAuthServiceImpl) GetOrCreateUser(ctx context.Context, giteeUser *GiteeUserInfo) (int64, string, error) {
	// Try to find user by gitee_id
	user, err := s.userRepo.FindByGiteeID(ctx, fmt.Sprintf("%d", giteeUser.ID))
	if err == nil && user != nil {
		// User exists, create session
		sessionKey, err := s.createSession(ctx, user.UserID)
		if err != nil {
			return 0, "", err
		}
		return user.UserID, sessionKey, nil
	}

	// Check if email already exists
	if giteeUser.Email != "" {
		existingUser, err := s.userRepo.GetUserByEmail(ctx, giteeUser.Email)
		if err == nil && existingUser != nil {
			// Email exists but not linked to Gitee, we need to link it
			err = s.userRepo.UpdateGiteeID(ctx, existingUser.UserID, fmt.Sprintf("%d", giteeUser.ID))
			if err != nil {
				return 0, "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("link gitee_id failed"))
			}
			sessionKey, err := s.createSession(ctx, existingUser.UserID)
			if err != nil {
				return 0, "", err
			}
			return existingUser.UserID, sessionKey, nil
		}
	}

	// Create new user
	hashedPassword, err := pswd.HashPassword(fmt.Sprintf("gitee_%d_%s", giteeUser.ID, time.Now().String()))
	if err != nil {
		return 0, "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("hash password failed"))
	}

	nickName := giteeUser.Name
	if nickName == "" {
		nickName = giteeUser.Login
	}

	newUser := &entity.User{
		UserID:       0,
		UniqueName:   "",
		NickName:     nickName,
		Email:        giteeUser.Email,
		HashPassword: hashedPassword,
		Description:  fmt.Sprintf("Gitee user: %s", giteeUser.Login),
		IconURI:      giteeUser.AvatarURL,
		GiteeID:     fmt.Sprintf("%d", giteeUser.ID),
	}

	userID, err := s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return 0, "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("create user failed"))
	}

	sessionKey, err := s.createSession(ctx, userID)
	if err != nil {
		return 0, "", err
	}

	return userID, sessionKey, nil
}

func (s *GiteeOAuthServiceImpl) createSession(ctx context.Context, userID int64) (string, error) {
	uniqueSessionID, err := s.idgen.GenID(ctx)
	if err != nil {
		return "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("generate session id failed"))
	}

	sessionDO := &session.Session{
		UserID:    fmt.Sprintf("%d", userID),
		SessionID: uniqueSessionID,
	}

	sessionKey, err := session.NewSessionService().GenerateSessionKey(ctx, sessionDO)
	if err != nil {
		return "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("generate session key failed"))
	}

	err = s.userRepo.UpdateSessionKey(ctx, userID, sessionKey)
	if err != nil {
		return "", errorx.WrapByCode(err, errno.CommonInternalErrorCode, errorx.WithExtraMsg("update session key failed"))
	}

	return sessionKey, nil
}
