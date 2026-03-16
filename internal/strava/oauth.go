package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAuthURL  = "https://www.strava.com/oauth/authorize"
	defaultTokenURL = "https://www.strava.com/oauth/token"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type OAuthClient struct {
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	EncryptionKey []byte

	tokenURL   string
	httpClient *http.Client
}

func NewOAuthClient(clientID, clientSecret, redirectURI string, encryptionKey []byte) *OAuthClient {
	return &OAuthClient{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirectURI,
		EncryptionKey: encryptionKey,
		tokenURL:      defaultTokenURL,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *OAuthClient) AuthURL() string {
	v := url.Values{}
	v.Set("client_id", c.ClientID)
	v.Set("redirect_uri", c.RedirectURI)
	v.Set("response_type", "code")
	v.Set("scope", "activity:read_all")
	return defaultAuthURL + "?" + v.Encode()
}

type stravaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type stravaErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors"`
}

func (c *OAuthClient) Exchange(ctx context.Context, code string) (*TokenPair, error) {
	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")

	return c.doTokenRequest(ctx, form)
}

func (c *OAuthClient) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")

	return c.doTokenRequest(ctx, form)
}

func (c *OAuthClient) IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

func (c *OAuthClient) EncryptToken(token string) ([]byte, error) {
	return Encrypt([]byte(token), c.EncryptionKey)
}

func (c *OAuthClient) DecryptToken(ciphertext []byte) (string, error) {
	plaintext, err := Decrypt(ciphertext, c.EncryptionKey)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (c *OAuthClient) doTokenRequest(ctx context.Context, form url.Values) (*TokenPair, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp stravaErrorResponse
		if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Message != "" {
			return nil, fmt.Errorf("strava API error (HTTP %d): %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("strava API error: HTTP %d", resp.StatusCode)
	}

	var tokenResp stravaTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	return &TokenPair{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Unix(tokenResp.ExpiresAt, 0),
	}, nil
}
