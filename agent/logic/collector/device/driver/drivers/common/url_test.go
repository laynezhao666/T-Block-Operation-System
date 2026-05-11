package common_test

import (
	"testing"

	"agent/logic/collector/device/driver/drivers/common"
)

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantProto string
		wantHost  string
		wantPort  int
		wantErr   bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:      "simple host",
			input:     "api.example.com",
			wantProto: "http",
			wantHost:  "api.example.com",
			wantPort:  0,
			wantErr:   false,
		},
		{
			name:      "host with port",
			input:     "api.example.com:8080",
			wantProto: "http",
			wantHost:  "api.example.com",
			wantPort:  8080,
			wantErr:   false,
		},
		{
			name:      "https scheme with host",
			input:     "https://api.example.com",
			wantProto: "https",
			wantHost:  "api.example.com",
			wantPort:  0,
			wantErr:   false,
		},
		{
			name:      "https scheme with port",
			input:     "https://api.example.com:443",
			wantProto: "https",
			wantHost:  "api.example.com",
			wantPort:  443,
			wantErr:   false,
		},
		{
			name:      "http scheme with port",
			input:     "http://localhost:3000",
			wantProto: "http",
			wantHost:  "localhost",
			wantPort:  3000,
			wantErr:   false,
		},
		{
			name:    "invalid port - negative",
			input:   "api.example.com:-1",
			wantErr: true,
		},
		{
			name:    "invalid port - too large",
			input:   "api.example.com:70000",
			wantErr: true,
		},
		{
			name:    "invalid port - non-numeric",
			input:   "api.example.com:abc",
			wantErr: true,
		},
		{
			name:      "uppercase scheme",
			input:     "HTTPS://api.example.com",
			wantProto: "https",
			wantHost:  "api.example.com",
			wantPort:  0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, host, port, err := common.ParseHostPort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHostPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proto != tt.wantProto {
					t.Errorf("ParseHostPort() proto = %v, want %v", proto, tt.wantProto)
				}
				if host != tt.wantHost {
					t.Errorf("ParseHostPort() host = %v, want %v", host, tt.wantHost)
				}
				if port != tt.wantPort {
					t.Errorf("ParseHostPort() port = %v, want %v", port, tt.wantPort)
				}
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		proto    string
		host     string
		port     int
		path     string
		expected string
	}{
		{
			name:     "simple http",
			proto:    "http",
			host:     "api.example.com",
			port:     0,
			path:     "/api/v1/data",
			expected: "http://api.example.com/api/v1/data",
		},
		{
			name:     "https with port",
			proto:    "https",
			host:     "api.example.com",
			port:     443,
			path:     "/api/v1/data",
			expected: "https://api.example.com:443/api/v1/data",
		},
		{
			name:     "path without leading slash",
			proto:    "http",
			host:     "localhost",
			port:     8080,
			path:     "api/data",
			expected: "http://localhost:8080/api/data",
		},
		{
			name:     "url encoded path",
			proto:    "https",
			host:     "api.example.com",
			port:     0,
			path:     "/api%2Fv1%2Fdata",
			expected: "https://api.example.com/api/v1/data",
		},
		{
			name:     "path with query params",
			proto:    "http",
			host:     "api.example.com",
			port:     0,
			path:     "/api/data?offset=0&limit=100",
			expected: "http://api.example.com/api/data?offset=0&limit=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.BuildURL(tt.proto, tt.host, tt.port, tt.path)
			if result != tt.expected {
				t.Errorf("BuildURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildURLWithOffset(t *testing.T) {
	tests := []struct {
		name      string
		urlStr    string
		newOffset int
		expected  string
	}{
		{
			name:      "update existing offset",
			urlStr:    "https://api.example.com/data?offset=0&limit=100",
			newOffset: 100,
			expected:  "https://api.example.com/data?limit=100&offset=100",
		},
		{
			name:      "add offset to url without offset",
			urlStr:    "https://api.example.com/data?limit=100",
			newOffset: 50,
			expected:  "https://api.example.com/data?limit=100&offset=50",
		},
		{
			name:      "url without query params",
			urlStr:    "https://api.example.com/data",
			newOffset: 100,
			expected:  "https://api.example.com/data?offset=100",
		},
		{
			name:      "invalid url returns original",
			urlStr:    "://invalid",
			newOffset: 100,
			expected:  "://invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.BuildURLWithOffset(tt.urlStr, tt.newOffset)
			if result != tt.expected {
				t.Errorf("BuildURLWithOffset() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name          string
		urlStr        string
		wantOffset    int
		wantLimit     int
		wantHasPaging bool
	}{
		{
			name:          "valid pagination params",
			urlStr:        "https://api.example.com/data?offset=0&limit=100",
			wantOffset:    0,
			wantLimit:     100,
			wantHasPaging: true,
		},
		{
			name:          "offset only",
			urlStr:        "https://api.example.com/data?offset=50",
			wantOffset:    0,
			wantLimit:     0,
			wantHasPaging: false,
		},
		{
			name:          "limit only",
			urlStr:        "https://api.example.com/data?limit=100",
			wantOffset:    0,
			wantLimit:     0,
			wantHasPaging: false,
		},
		{
			name:          "no pagination params",
			urlStr:        "https://api.example.com/data",
			wantOffset:    0,
			wantLimit:     0,
			wantHasPaging: false,
		},
		{
			name:          "non-numeric offset",
			urlStr:        "https://api.example.com/data?offset=abc&limit=100",
			wantOffset:    0,
			wantLimit:     0,
			wantHasPaging: false,
		},
		{
			name:          "non-numeric limit",
			urlStr:        "https://api.example.com/data?offset=0&limit=abc",
			wantOffset:    0,
			wantLimit:     0,
			wantHasPaging: false,
		},
		{
			name:          "large offset and limit",
			urlStr:        "https://api.example.com/data?offset=10000&limit=500",
			wantOffset:    10000,
			wantLimit:     500,
			wantHasPaging: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, limit, hasPaging := common.ParsePaginationParams(tt.urlStr)
			if offset != tt.wantOffset {
				t.Errorf("ParsePaginationParams() offset = %v, want %v", offset, tt.wantOffset)
			}
			if limit != tt.wantLimit {
				t.Errorf("ParsePaginationParams() limit = %v, want %v", limit, tt.wantLimit)
			}
			if hasPaging != tt.wantHasPaging {
				t.Errorf("ParsePaginationParams() hasPaging = %v, want %v", hasPaging, tt.wantHasPaging)
			}
		})
	}
}
