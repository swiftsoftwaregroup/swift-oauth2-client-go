// Package oauth2client provides a client for making authenticated API calls using OAuth2.
//
// This package simplifies the process of making authenticated requests to APIs that use
// OAuth2 for authentication. It handles token acquisition, automatic token refresh,
// and provides methods for making various types of API calls including GET, POST,
// and file downloads.
//
// Usage:
//
//	config := oauth2client.OAuth2Config{
//		TokenURL:     "https://api.example.com/oauth/token",
//		ClientID:     "your_client_id",
//		ClientSecret: "your_client_secret",
//		Scopes:       []string{"read", "write"},
//	}
//	client := oauth2client.NewAPIClient(config, "https://api.example.com")
//
//	// Make a GET request
//	response, statusCode, err := client.CallAPI(oauth2client.HttpGet, "/users", nil, nil)
//
//	// Make a POST request
//	postBody := map[string]string{"name": "John Doe"}
//	response, statusCode, err := client.CallAPI(oauth2client.HttpPost, "/users", postBody, nil)
//
//	// Download a file
//	err := client.DownloadFile(oauth2client.HttpGet, "/files/document.pdf", nil, nil, "./document.pdf")
package oauth2client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// APIClient is a client for making authenticated API calls using OAuth2.
// It handles token management and provides methods for various types of API requests.
type APIClient struct {
	tokenManager *tokenManager
	baseURL      string
	httpClient   *http.Client
}

// NewAPIClient creates a new APIClient with the given OAuth2 configuration and base URL.
//
// Parameters:
//   - config: The OAuth2 configuration including token URL, client credentials, and scopes.
//   - baseURL: The base URL of the API you're accessing.
//
// Returns:
//   - *APIClient: A new instance of APIClient.
//
// Example:
//
//	config := oauth2client.OAuth2Config{
//		TokenURL:     "https://api.example.com/oauth/token",
//		ClientID:     "your_client_id",
//		ClientSecret: "your_client_secret",
//		Scopes:       []string{"read", "write"},
//	}
//	client := oauth2client.NewAPIClient(config, "https://api.example.com")
func NewAPIClient(config OAuth2Config, baseURL string) *APIClient {
	return &APIClient{
		tokenManager: &tokenManager{config: config},
		baseURL:      baseURL,
		httpClient:   &http.Client{},
	}
}

// CallAPI makes an authenticated API call and returns the response body, status code, and any error.
//
// Parameters:
//   - method: The HTTP method to use (e.g., HttpGet, HttpPost, HttpPut, HttpDelete)
//   - path: The API endpoint path (will be appended to the base URL)
//   - body: The request body. Can be nil, a string, []byte, url.Values, or any JSON-serializable type
//   - additionalHeaders: Additional HTTP headers to include in the request
//
// Returns:
//   - []byte: The response body
//   - int: The HTTP status code
//   - error: Any error that occurred during the request
//
// Example:
//
//	// Make a GET request
//	response, statusCode, err := client.CallAPI(oauth2client.HttpGet, "/users", nil, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Status: %d, Response: %s\n", statusCode, string(response))
//
//	// Make a POST request with JSON body
//	postBody := map[string]string{"name": "John Doe"}
//	response, statusCode, err := client.CallAPI(oauth2client.HttpPost, "/users", postBody, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Status: %d, Response: %s\n", statusCode, string(response))
func (c *APIClient) CallAPI(method HttpMethod, path string, body interface{}, additionalHeaders map[string]string) ([]byte, int, error) {
	token, err := c.tokenManager.getValidToken()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get valid token: %w", err)
	}

	var bodyReader io.Reader
	var contentType string

	switch v := body.(type) {
	case nil:
		// No body
	case string:
		bodyReader = strings.NewReader(v)
		contentType = "text/plain"
	case []byte:
		bodyReader = bytes.NewReader(v)
		contentType = "application/octet-stream"
	case url.Values:
		bodyReader = strings.NewReader(v.Encode())
		contentType = "application/x-www-form-urlencoded"
	default:
		jsonBody, err := json.Marshal(v)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		contentType = "application/json"
	}

	req, err := http.NewRequest(string(method), c.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, value := range additionalHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Token might have expired, try refreshing and calling again
		if err := c.tokenManager.refreshToken(); err != nil {
			return nil, 0, fmt.Errorf("failed to refresh token: %w", err)
		}
		return c.CallAPI(method, path, body, additionalHeaders) // Recursive call with fresh token
	}

	// Consider both 200 OK and 201 Created as successful responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, resp.StatusCode, fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, resp.StatusCode, nil
}

// DownloadFile downloads a file from the specified API endpoint and saves it to the given destination path.
//
// Parameters:
//   - method: The HTTP method to use (typically HttpGet)
//   - path: The API endpoint path for the file download
//   - body: The request body (if any). Can be nil, a string, []byte, url.Values, or any JSON-serializable type
//   - additionalHeaders: Additional HTTP headers to include in the request
//   - destPath: The local file path where the downloaded file should be saved
//
// Returns:
//   - error: Any error that occurred during the download process
//
// Example:
//
//	err := client.DownloadFile(oauth2client.HttpGet, "/files/document.pdf", nil, nil, "./downloaded_document.pdf")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("File downloaded successfully")
func (c *APIClient) DownloadFile(method HttpMethod, path string, body interface{}, additionalHeaders map[string]string, destPath string) error {
	token, err := c.tokenManager.getValidToken()
	if err != nil {
		return fmt.Errorf("failed to get valid token: %w", err)
	}

	var bodyReader io.Reader
	var contentType string

	switch v := body.(type) {
	case nil:
		// No body
	case string:
		bodyReader = strings.NewReader(v)
		contentType = "text/plain"
	case []byte:
		bodyReader = bytes.NewReader(v)
		contentType = "application/octet-stream"
	case url.Values:
		bodyReader = strings.NewReader(v.Encode())
		contentType = "application/x-www-form-urlencoded"
	default:
		jsonBody, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		contentType = "application/json"
	}

	req, err := http.NewRequest(string(method), c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, value := range additionalHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.tokenManager.refreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		return c.DownloadFile(method, path, body, additionalHeaders, destPath)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// If destPath is a directory, try to get filename from Content-Disposition header
	if fi, err := os.Stat(destPath); err == nil && fi.IsDir() {
		if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
			if _, params, err := mime.ParseMediaType(disposition); err == nil {
				if filename, ok := params["filename"]; ok {
					destPath = filepath.Join(destPath, filename)
				}
			}
		}
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}