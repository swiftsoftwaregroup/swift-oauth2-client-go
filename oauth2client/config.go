// Package oauth2client provides a client for making authenticated API calls using OAuth2.
package oauth2client

// OAuth2Config holds the configuration for OAuth2 authentication.
type OAuth2Config struct {
	// TokenURL is the URL of the token endpoint.
	TokenURL string

	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// Scopes is a list of requested permission scopes.
	Scopes []string
}

// tokenResponse represents the server's response to a token request.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}
