package oauth2client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAPIClient(t *testing.T) {
	// Mock OAuth2 token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/token" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test_access_token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer tokenServer.Close()

	// Mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Unexpected Authorization header: %s", r.Header.Get("Authorization"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/api/test":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "success"})
		case "/api/download":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="test.txt"`)
			w.Write([]byte("test file content"))
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	// Create APIClient
	config := OAuth2Config{
		TokenURL:     tokenServer.URL + "/token",
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Scopes:       []string{"test_scope"},
	}
	client := NewAPIClient(&config, apiServer.URL)

	// Test CallAPI with GET
	t.Run("CallAPI GET", func(t *testing.T) {
		response, statusCode, err := client.CallAPI(HttpGet, "/api/test", nil, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Errorf("Unexpected status code: %d", statusCode)
		}

		var result map[string]string
		if err := json.Unmarshal(response, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if result["message"] != "success" {
			t.Errorf("Unexpected response: %v", result)
		}
	})

	// Test CallAPI with POST
	t.Run("CallAPI POST", func(t *testing.T) {
		postBody := map[string]string{"key": "value"}
		response, statusCode, err := client.CallAPI(HttpPost, "/api/test", postBody, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Errorf("Unexpected status code: %d", statusCode)
		}

		var result map[string]string
		if err := json.Unmarshal(response, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if result["message"] != "success" {
			t.Errorf("Unexpected response: %v", result)
		}
	})

	// Test DownloadFile
	t.Run("DownloadFile", func(t *testing.T) {
		tempFile := t.TempDir() + "/test_download.txt"
		err := client.DownloadFile(HttpGet, "/api/download", nil, nil, tempFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}
		if string(content) != "test file content" {
			t.Errorf("Unexpected file content: %s", string(content))
		}
	})

	// Test token refresh
	t.Run("Token Refresh", func(t *testing.T) {
		// Force token expiration
		client.tokenManager.expiresAt = time.Now().Add(-1 * time.Hour)

		response, statusCode, err := client.CallAPI(HttpGet, "/api/test", nil, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Errorf("Unexpected status code: %d", statusCode)
		}

		var result map[string]string
		if err := json.Unmarshal(response, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if result["message"] != "success" {
			t.Errorf("Unexpected response: %v", result)
		}
	})

	// Test CallAPI with GET and gzip encoding
	t.Run("CallAPI GET with gzip encoding", func(t *testing.T) {
		gzipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", "application/json")
			gz := gzip.NewWriter(w)
			json.NewEncoder(gz).Encode(map[string]string{"message": "gzip success"})
			gz.Close()
		}))
		defer gzipServer.Close()

		gzipClient := NewAPIClient(nil, gzipServer.URL) // Using it as a thin wrapper without OAuth

		response, statusCode, err := gzipClient.CallAPI(HttpGet, "/", nil, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Errorf("Unexpected status code: %d", statusCode)
		}

		var result map[string]string
		if err := json.Unmarshal(response, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if result["message"] != "gzip success" {
			t.Errorf("Unexpected response: %v", result)
		}
	})

	t.Run("CallAPIWithContext-Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, statusCode, err := client.CallAPIWithContext(ctx, HttpGet, "/api/test", nil, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Errorf("Unexpected status code: %d", statusCode)
		}

		var result map[string]string
		if err := json.Unmarshal(response, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if result["message"] != "success" {
			t.Errorf("Unexpected response: %v", result)
		}
	})

	t.Run("CallAPIWithContext-Timeout", func(t *testing.T) {
		// Create a server with artificial delay
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer slowServer.Close()

		slowClient := NewAPIClient(nil, slowServer.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, _, err := slowClient.CallAPIWithContext(ctx, HttpGet, "/", nil, nil)
		if err == nil {
			t.Fatal("Expected timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected deadline exceeded error, got: %v", err)
		}
	})

	t.Run("DownloadFileWithContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tempFile := t.TempDir() + "/test_download_with_context.txt"
		err := client.DownloadFileWithContext(ctx, HttpGet, "/api/download", nil, nil, tempFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}
		if string(content) != "test file content" {
			t.Errorf("Unexpected file content: %s", string(content))
		}
	})
}
