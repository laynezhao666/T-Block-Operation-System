// Package common provides reusable utilities for HTTP-based drivers
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	// DefaultRefreshTokenBuffer is the minimum seconds before token expiry to refresh
	DefaultRefreshTokenBuffer = 600
)

// OAuth2Config holds OAuth2 token configuration
type OAuth2Config struct {
	// Token request configuration
	ReqMethod string // Request method (default: POST)
	ReqURL    string // Token request URL
	ReqBody   string // Request body

	// Token response parsing configuration
	RespTokenKey   string // Field name for token in response (default: access_token)
	RespTimeoutKey string // Field name for timeout in response (default: token_timeout)

	// Header configuration for subsequent requests
	HeaderName  string // Header name (default: Authorization)
	HeaderValue string // Header value format (default: Bearer {access_token})
}

// OAuth2Manager manages OAuth2 token lifecycle
type OAuth2Manager struct {
	config OAuth2Config
	client *http.Client
	token  atomic.String
	header map[string]string

	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
	deviceName string // For logging purposes
}

// NewOAuth2Manager creates a new OAuth2 token manager
// Parameters:
// - client: HTTP client for token requests
// - deviceName: device name for logging
// Returns:
// - *OAuth2Manager: new manager instance
func NewOAuth2Manager(client *http.Client, deviceName string) *OAuth2Manager {
	return &OAuth2Manager{
		client:     client,
		deviceName: deviceName,
		header:     make(map[string]string),
	}
}

// Initialize initializes the OAuth2 manager with the given configuration
// Parameters:
// - ctx: context for initialization
// - config: OAuth2 configuration
// Returns:
// - error: initialization error
func (m *OAuth2Manager) Initialize(ctx context.Context, config OAuth2Config) error {
	m.config = config

	// Set defaults
	if m.config.ReqMethod == "" {
		m.config.ReqMethod = "POST"
	}
	if m.config.RespTokenKey == "" {
		m.config.RespTokenKey = "access_token"
	}
	if m.config.RespTimeoutKey == "" {
		m.config.RespTimeoutKey = "token_timeout"
	}
	if m.config.HeaderName == "" {
		m.config.HeaderName = "Authorization"
	}
	if m.config.HeaderValue == "" {
		m.config.HeaderValue = fmt.Sprintf("Bearer %s", AccessTokenPlaceholder)
	}

	// Fetch initial token
	token, timeout, err := m.fetchToken(ctx)
	if err != nil {
		return err
	}

	m.token.Store(token)
	m.header = map[string]string{m.config.HeaderName: m.config.HeaderValue}

	// Start token refresh goroutine
	refreshCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runTokenRefresh(refreshCtx, timeout-DefaultRefreshTokenBuffer)
	}()

	return nil
}

// GetToken returns the current token
func (m *OAuth2Manager) GetToken() string {
	return m.token.Load()
}

// GetHeaders returns the configured headers with token placeholder
func (m *OAuth2Manager) GetHeaders() map[string]string {
	result := make(map[string]string)
	for k, v := range m.header {
		result[k] = v
	}
	return result
}

// RefreshToken manually refreshes the token
// Returns:
// - error: refresh error
func (m *OAuth2Manager) RefreshToken(ctx context.Context) error {
	if m.config.ReqURL == "" {
		return nil
	}

	token, _, err := m.fetchToken(ctx)
	if err != nil {
		return err
	}

	m.token.Store(token)
	log.Debugf("%s refresh token success", m.deviceName)
	return nil
}

// Stop stops the token refresh goroutine and waits for cleanup
func (m *OAuth2Manager) Stop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	m.wg.Wait()
	m.token.Store("")
	m.header = make(map[string]string)
}

// Wait waits for all goroutines to exit
func (m *OAuth2Manager) Wait() {
	m.wg.Wait()
}

// fetchToken fetches a new token from the OAuth2 server
func (m *OAuth2Manager) fetchToken(ctx context.Context) (string, int, error) {
	req, err := m.createTokenRequest(ctx)
	if err != nil {
		return "", 0, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		log.Warnf("Error sending request: %v, url:%s", err, m.config.ReqURL)
		return "", 0, err
	}
	defer resp.Body.Close()

	return m.parseTokenResponse(resp.Body)
}

// createTokenRequest creates a new token request
func (m *OAuth2Manager) createTokenRequest(ctx context.Context) (*http.Request, error) {
	var req *http.Request
	var err error

	if m.config.ReqMethod == "POST" {
		req, err = http.NewRequestWithContext(ctx, m.config.ReqMethod, m.config.ReqURL,
			strings.NewReader(m.config.ReqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, m.config.ReqMethod, m.config.ReqURL, nil)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

// parseTokenResponse parses the token response
func (m *OAuth2Manager) parseTokenResponse(body io.ReadCloser) (string, int, error) {
	str, err := io.ReadAll(body)
	if err != nil {
		log.Warnf("Error reading response: %v", err)
		return "", 0, err
	}

	out := make(map[string]any)
	if err := json.Unmarshal(str, &out); err != nil {
		log.Warnf("Error decoding response: %v, %s", err, str)
		return "", 0, err
	}

	token, has := out[m.config.RespTokenKey]
	if !has {
		log.Warnf("token not found:%+v", out)
		return "", 0, fmt.Errorf("token not found")
	}

	timeout, has := out[m.config.RespTimeoutKey]
	if !has {
		log.Warnf("token_timeout not found:%+v, key:%s", out, m.config.RespTimeoutKey)
		return "", 0, fmt.Errorf("token_timeout not found")
	}

	timeoutInt, err := strconv.Atoi(fmt.Sprintf("%v", timeout))
	if err != nil {
		return "", 0, fmt.Errorf("invalid token_timeout: %v", timeout)
	}

	return fmt.Sprintf("%v", token), timeoutInt, nil
}

// runTokenRefresh runs the token refresh loop
func (m *OAuth2Manager) runTokenRefresh(ctx context.Context, timeout int) {
	if timeout <= 0 {
		timeout = DefaultRefreshTokenBuffer
	}

	interval := time.Duration(timeout) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debugf("Token update goroutine stopped due to context cancellation")
			return
		case <-ticker.C:
			m.doRefreshToken(ctx, &ticker, &interval, &timeout)
		}
	}
}

// doRefreshToken performs the actual token refresh
func (m *OAuth2Manager) doRefreshToken(ctx context.Context, ticker **time.Ticker, interval *time.Duration, timeout *int) {
	var resp *http.Response
	var err error
	now := time.Now()
	try := 0

	for {
		req, reqErr := m.createTokenRequest(ctx)
		if reqErr != nil {
			log.Warnf("Error creating token request: %v", reqErr)
			break
		}
		resp, err = m.client.Do(req)
		if err == nil || time.Since(now) >= *interval {
			break
		}
		try++
		time.Sleep(time.Second)
	}

	if err != nil {
		log.Warnf("Error sending request: %v, try:%d, url:%s", err, try, m.config.ReqURL)
		return
	}

	if resp == nil {
		return
	}
	defer resp.Body.Close()

	token, newTimeoutInt, errParse := m.parseTokenResponse(resp.Body)
	if errParse != nil {
		return
	}

	log.Debugf("%s update token %d", m.deviceName, newTimeoutInt)
	m.token.Store(token)

	// Adjust ticker if new timeout is shorter
	if newTimeoutInt-DefaultRefreshTokenBuffer < *timeout {
		newInterval := time.Duration(newTimeoutInt-DefaultRefreshTokenBuffer) * time.Second
		if newInterval > 0 {
			log.Debugf("%s adjusting token update interval to %v", m.deviceName, newInterval)
			(*ticker).Reset(newInterval)
			*interval = newInterval
			*timeout = newTimeoutInt - DefaultRefreshTokenBuffer
		}
	}
}
