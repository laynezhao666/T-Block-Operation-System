// Package common provides reusable utilities for HTTP-based drivers
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// AccessTokenPlaceholder is the placeholder for access token in URLs and headers
	AccessTokenPlaceholder = "{access_token}"
)

// DoJSON executes an HTTP request with JSON body and response
// Parameters:
// - ctx: context for request cancellation
// - client: HTTP client
// - method: HTTP method (GET, POST, PUT, DELETE)
// - urlStr: URL string
// - body: request body (for POST/PUT)
// - headers: request headers
// - token: bearer token (optional, replaces placeholder in URL and headers)
// - out: output variable for JSON response
// Returns:
// - error: request error
func DoJSON(
	ctx context.Context,
	client *http.Client,
	method, urlStr string,
	body map[string]any,
	headers map[string]string,
	token string,
	out any,
) error {
	var req *http.Request
	var err error

	// Replace token placeholder in URL
	urlReal := urlStr
	tokenTrimmed := strings.TrimSpace(token)
	if tokenTrimmed != "" {
		urlReal = strings.ReplaceAll(urlStr, AccessTokenPlaceholder, tokenTrimmed)
	}

	// Determine method if not specified
	if len(method) == 0 {
		if body != nil && len(body) > 0 {
			method = "POST"
		} else {
			method = "GET"
		}
	}

	switch strings.ToUpper(method) {
	case "GET", "DELETE":
		req, err = http.NewRequestWithContext(ctx, method, urlReal, nil)
	case "POST", "PUT":
		var b []byte
		if body != nil {
			b, err = json.Marshal(body)
			if err != nil {
				return err
			}
		} else {
			b = []byte("{}")
		}
		req, err = http.NewRequestWithContext(ctx, method, urlReal, strings.NewReader(string(b)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return err
	}

	// Set headers with token replacement
	for k, v := range headers {
		req.Header.Set(k, strings.ReplaceAll(v, AccessTokenPlaceholder, tokenTrimmed))
	}

	if client == nil {
		return fmt.Errorf("HTTP client is not initialized")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check HTTP response status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP request failed with status %d and failed to read response body: %v",
				resp.StatusCode, readErr)
		}
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	str, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if len(str) == 0 {
		return fmt.Errorf("empty response body")
	}

	err = json.Unmarshal(str, out)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %v, data: %s", err, string(str))
	}

	return nil
}

// DoJSONRaw executes an HTTP request with raw string body and JSON response
// Parameters:
// - ctx: context for request cancellation
// - client: HTTP client
// - method: HTTP method
// - urlStr: URL string
// - bodyStr: raw request body string
// - contentType: content type header (defaults to application/json if empty)
// - out: output variable for JSON response
// Returns:
// - error: request error
func DoJSONRaw(
	ctx context.Context,
	client *http.Client,
	method, urlStr string,
	bodyStr string,
	contentType string,
	out any,
) error {
	var req *http.Request
	var err error

	if method == "POST" || method == "PUT" {
		req, err = http.NewRequestWithContext(ctx, method, urlStr, strings.NewReader(bodyStr))
		if err != nil {
			return err
		}
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, urlStr, nil)
	}

	if err != nil {
		return err
	}

	if client == nil {
		return fmt.Errorf("HTTP client is not initialized")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	str, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if len(str) == 0 {
		return fmt.Errorf("empty response body")
	}

	err = json.Unmarshal(str, out)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %v, data: %s", err, string(str))
	}

	return nil
}
