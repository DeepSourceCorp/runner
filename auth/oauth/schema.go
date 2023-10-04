package oauth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type AuthorizationRequest struct {
	AppID    string `param:"app_id"`
	ClientID string `query:"client_id"`
	Scopes   string `query:"scopes"`
	State    string `query:"state"`
}

func NewAuthorizationRequest(c echo.Context) (*AuthorizationRequest, error) {
	req := &AuthorizationRequest{}
	if err := c.Bind(req); err != nil {
		err = fmt.Errorf("authorize request bind error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *AuthorizationRequest) Validate() error {
	if r.AppID == "" || r.ClientID == "" || r.State == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type CallbackRequest struct {
	AppID string `param:"app_id"`
	Code  string `query:"code"`
	State string `query:"state"`
}

func NewCallbackRequest(c echo.Context) (*CallbackRequest, error) {
	req := &CallbackRequest{}
	if err := c.Bind(req); err != nil {
		err = fmt.Errorf("callback request bind error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *CallbackRequest) Validate() error {
	if r.AppID == "" || r.Code == "" || r.State == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type SessionRequest struct {
	AppID        string `param:"app_id"`
	State        string `query:"state"`
	SessionToken string
}

func NewSessionRequest(c echo.Context) (*SessionRequest, error) {
	req := &SessionRequest{}
	if err := c.Bind(req); err != nil {
		err = fmt.Errorf("session request bind error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}

	cookie, err := c.Cookie("session")
	if err != nil {
		err = fmt.Errorf("session request cookie error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}

	sessionToken := cookie.Value
	req.SessionToken = sessionToken

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *SessionRequest) Validate() error {
	if r.AppID == "" || r.State == "" || r.SessionToken == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type TokenRequest struct {
	AppID        string `param:"app_id"`
	Code         string `query:"code" json:"code"`
	ClientID     string `query:"client_id" json:"client_id"`
	ClientSecret string `query:"client_secret" json:"client_secret"`
}

func NewTokenRequest(c echo.Context) (*TokenRequest, error) {
	req := &TokenRequest{}
	if err := c.Bind(req); err != nil {
		err = fmt.Errorf("token request bind error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *TokenRequest) Validate() error {
	if r.AppID == "" || r.Code == "" || r.ClientID == "" || r.ClientSecret == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Expiry       time.Time `json:"expiry"`
	RefreshToken string    `json:"refresh_token"`
}

func NewTokenResponse(token *oauth2.Token) *TokenResponse {
	expiresIn := int(time.Until(token.Expiry).Seconds())
	return &TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		ExpiresIn:    expiresIn,
		RefreshToken: token.RefreshToken,
	}
}

type UserRequest struct {
	AppID       string `param:"app_id"`
	AccessToken string
}

func NewUserRequest(c echo.Context) (*UserRequest, error) {
	req := &UserRequest{}
	req.AppID = c.Param("app_id")

	authorization := c.Request().Header.Get("Authorization")
	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, httperror.ErrBadRequest(errors.New("invalid authorization header"))
	}

	req.AccessToken = parts[1]

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *UserRequest) Validate() error {
	if r.AppID == "" || r.AccessToken == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type UserResponse struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	ID       string `json:"id"`
}

type RefreshRequest struct {
	AppID        string `param:"app_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshRequest(c echo.Context) (*RefreshRequest, error) {
	req := &RefreshRequest{}
	if err := c.Bind(req); err != nil {
		err = fmt.Errorf("refresh request bind error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *RefreshRequest) Validate() error {
	if r.AppID == "" || r.ClientID == "" || r.ClientSecret == "" || r.RefreshToken == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}
