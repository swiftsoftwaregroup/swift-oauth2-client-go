package oauth2client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// tokenManager handles OAuth2 token acquisition and refresh.
type tokenManager struct {
	config      OAuth2Config
	accessToken string
	expiresAt   time.Time
	mutex       sync.Mutex
}

// getValidToken returns a valid access token, refreshing if necessary.
func (tm *tokenManager) getValidToken() (string, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.accessToken == "" || time.Now().After(tm.expiresAt) {
		if err := tm.refreshToken(); err != nil {
			return "", err
		}
	}
	return tm.accessToken, nil
}

// refreshToken requests a new access token from the authorization server.
func (tm *tokenManager) refreshToken() error {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", tm.config.ClientID, tm.config.ClientSecret)))
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", strings.Join(tm.config.Scopes, " "))

	req, err := http.NewRequest("POST", tm.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get token: %s", string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	tm.accessToken = tokenResp.AccessToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	return nil
}
