package oauth2client

// HttpMethod represents an HTTP method.
type HttpMethod string

// HTTP Method constants
const (
	HttpGet     HttpMethod = "GET"
	HttpPost    HttpMethod = "POST"
	HttpPut     HttpMethod = "PUT"
	HttpDelete  HttpMethod = "DELETE"
	HttpPatch   HttpMethod = "PATCH"
	HttpHead    HttpMethod = "HEAD"
	HttpOptions HttpMethod = "OPTIONS"
)
