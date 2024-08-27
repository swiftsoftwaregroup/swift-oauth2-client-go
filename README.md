# Swift OAuth2 Client for Go

This SDK, developed by Swift Software Group, provides a simple and efficient way to make authenticated API calls using OAuth2 in Go applications. It handles token acquisition, automatic token refresh, and provides methods for making various types of API calls including GET, POST, and file downloads.

## Installation

To install the Swift OAuth2 Client Go SDK, use `go get`:

```bash
go get github.com/swiftsoftwaregroup/swift-oauth2-client-go
```

## Usage

Here's a quick example of how to use the Swift OAuth2 Client for Go:

```go
package main

import (
    "fmt"
    "log"

    "github.com/swiftsoftwaregroup/swift-oauth2-client-go/oauth2client"
)

func main() {
    config := oauth2client.OAuth2Config{
        TokenURL:     "https://api.example.com/oauth/token",
        ClientID:     "your_client_id",
        ClientSecret: "your_client_secret",
        Scopes:       []string{"read", "write"},
    }
    client := oauth2client.NewAPIClient(config, "https://api.example.com")

    // Make a GET request
    response, statusCode, err := client.CallAPI(oauth2client.HttpGet, "/users", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("GET Status: %d, Response: %s\n", statusCode, string(response))

    // Make a POST request
    postBody := map[string]string{"name": "John Doe"}
    response, statusCode, err = client.CallAPI(oauth2client.HttpPost, "/users", postBody, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("POST Status: %d, Response: %s\n", statusCode, string(response))

    // Download a file
    err = client.DownloadFile(oauth2client.HttpGet, "/files/document.pdf", nil, nil, "./document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("File downloaded successfully")
}
```

For more detailed examples, please check the `examples` directory in this repository.

## Documentation

For full documentation, please refer to the [GoDoc](https://pkg.go.dev/github.com/swiftsoftwaregroup/swift-oauth2-client-go/oauth2client) page.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Testing

```bash
go test -v ./...
```

## License

This SDK is distributed under the Apache 2.0 license. See the LICENSE file for more information.