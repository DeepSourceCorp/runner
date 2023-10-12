package oauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

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

	req.SessionToken = cookie.Value

	if err := req.validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *SessionRequest) validate() error {
	if r.AppID == "" || r.State == "" || r.SessionToken == "" {
		return errors.New("session request validation failed")
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
	if err := req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *TokenRequest) validate() error {
	if r.AppID == "" || r.Code == "" || r.ClientID == "" || r.ClientSecret == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	Expiry       time.Time `json:"expiry"`
	RefreshToken string    `json:"refresh_token"`
}

func NewTokenResponse(token *oauth2.Token) *TokenResponse {
	expiresIn := int64(time.Until(token.Expiry).Seconds())
	return &TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		ExpiresIn:    expiresIn,
		RefreshToken: token.RefreshToken,
	}
}
