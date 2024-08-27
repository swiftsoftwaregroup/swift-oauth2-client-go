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
