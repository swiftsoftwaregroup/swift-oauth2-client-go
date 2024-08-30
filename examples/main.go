package main

import (
	"fmt"
	"log"
	"os"

	"github.com/swiftsoftwaregroup/swift-oauth2-client-go/oauth2client"
)

func main() {
	config := oauth2client.OAuth2Config{
		TokenURL:     "http://localhost:5000/token",
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Scopes:       []string{"api:read", "api:write"},
	}

	client := oauth2client.NewAPIClient(&config, "http://localhost:5001")

	// GET request to /api/protected
	fmt.Println("Calling protected API with GET...")
	result, statusCode, err := client.CallAPI(oauth2client.HttpGet, "/api/protected", nil, nil)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("GET response (Status %d): %s\n", statusCode, string(result))

	// POST request to /api/resource
	fmt.Println("\nCalling resource API with POST...")
	postBody := map[string]string{"name": "example resource"}
	additionalHeaders := map[string]string{
		"Content-Type": "application/json",
	}
	result, statusCode, err = client.CallAPI(oauth2client.HttpPost, "/api/resource", postBody, additionalHeaders)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("POST response (Status %d): %s\n", statusCode, string(result))

	// File download from /api/download
	fmt.Println("\nDownloading file...")
	downloadPath := "./example.txt"
	err = client.DownloadFile(oauth2client.HttpGet, "/api/download", nil, nil, downloadPath)
	if err != nil {
		log.Fatalf("Error downloading file: %v\n", err)
	}
	fmt.Println("File downloaded successfully")

	// Read and print the content of the downloaded file
	content, err := os.ReadFile(downloadPath)
	if err != nil {
		log.Fatalf("Error reading downloaded file: %v\n", err)
	}
	fmt.Printf("Downloaded file content: %s\n", string(content))

	// Clean up the downloaded file
	os.Remove(downloadPath)
}
