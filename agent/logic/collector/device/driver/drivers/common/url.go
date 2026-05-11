// Package common provides reusable utilities for HTTP-based drivers
package common

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParseHostPort parses host and port from a URL string
// Parameters:
// - name: URL string, may contain protocol prefix and port
// Returns:
// - proto: protocol (http/https)
// - host: host address
// - port: port number (0 if not specified)
// - error: parse error
func ParseHostPort(name string) (proto string, host string, port int, err error) {
	proto = "http"
	s := strings.TrimSpace(name)
	if s == "" {
		return "", "", 0, errors.New("empty host")
	}

	// Remove scheme
	if i := strings.Index(s, "//"); i >= 0 {
		if i > len(proto) {
			proto = strings.ToLower(s[:i-1])
		}
		s = s[i+2:]
	}

	// Parse port
	host = s
	port = 0
	if j := strings.LastIndex(s, ":"); j > 0 {
		host = s[:j]
		pstr := s[j+1:]
		p, parseErr := strconv.Atoi(pstr)
		if parseErr != nil || p <= 0 || p > 65535 {
			return "", "", 0, fmt.Errorf("invalid port: %s", pstr)
		}
		port = p
	}
	return proto, host, port, nil
}

// BuildURL constructs a full URL from components
// Parameters:
// - proto: protocol (http/https)
// - host: host address
// - port: port number (0 to omit)
// - path: path portion
// Returns:
// - string: complete URL
func BuildURL(proto, host string, port int, path string) string {
	var tmp string
	if strings.HasPrefix(path, "/") {
		tmp = path
	} else {
		tmp = "/" + path
	}

	pathNew, err := url.QueryUnescape(tmp)
	if err != nil {
		pathNew = tmp
	}

	if port > 0 {
		return fmt.Sprintf("%s://%s:%d%s", proto, host, port, pathNew)
	}
	return fmt.Sprintf("%s://%s%s", proto, host, pathNew)
}

// BuildURLWithOffset constructs a URL with new offset parameter for pagination
// Parameters:
// - urlStr: original URL string
// - newOffset: new offset value
// Returns:
// - string: URL with updated offset
func BuildURLWithOffset(urlStr string, newOffset int) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	query := parsedURL.Query()
	query.Set("offset", strconv.Itoa(newOffset))
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

// ParsePaginationParams extracts offset and limit from URL query parameters
// Parameters:
// - urlStr: URL string
// Returns:
// - offset: offset value
// - limit: limit value
// - hasPagination: whether URL contains pagination parameters
func ParsePaginationParams(urlStr string) (offset int, limit int, hasPagination bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return 0, 0, false
	}

	query := parsedURL.Query()
	offsetStr := query.Get("offset")
	limitStr := query.Get("limit")

	if offsetStr == "" || limitStr == "" {
		return 0, 0, false
	}

	offset, err1 := strconv.Atoi(offsetStr)
	limit, err2 := strconv.Atoi(limitStr)

	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	return offset, limit, true
}
