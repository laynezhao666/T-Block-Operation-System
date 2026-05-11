// Package common provides reusable utilities for HTTP-based drivers
package common

import (
	"time"
)

// BaseCacheItem represents a basic cache item with expiration
type BaseCacheItem struct {
	Value      string    // Cached value
	ExpireTime time.Time // Data expiration time
}

// IsExpired checks if the cache item has expired
func (c *BaseCacheItem) IsExpired() bool {
	return time.Now().After(c.ExpireTime)
}

// CacheState tracks cache update state for an API
type CacheState struct {
	FailureCount int       // Consecutive failure count
	LastUpdate   time.Time // Last successful update time
}

// IncrementFailure increments the failure count
func (s *CacheState) IncrementFailure() {
	s.FailureCount++
}

// ResetFailure resets the failure count and updates last update time
func (s *CacheState) ResetFailure() {
	s.FailureCount = 0
	s.LastUpdate = time.Now()
}

// IsHealthy checks if the cache state is healthy based on max failures threshold
func (s *CacheState) IsHealthy(maxFailures int) bool {
	return s.FailureCount < maxFailures
}

// PaginationInfo represents pagination information from API response
type PaginationInfo struct {
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
	Total    int    `json:"total"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
}

// ParsePaginationFromResponse parses pagination info from response data
// Parameters:
// - respData: response data map
// Returns:
// - *PaginationInfo: parsed pagination info, nil if not found
func ParsePaginationFromResponse(respData map[string]interface{}) *PaginationInfo {
	paginationRaw, ok := respData["pagination"]
	if !ok {
		return nil
	}

	paginationMap, ok := paginationRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	pagination := &PaginationInfo{}

	if offset, ok := paginationMap["offset"].(float64); ok {
		pagination.Offset = int(offset)
	}
	if limit, ok := paginationMap["limit"].(float64); ok {
		pagination.Limit = int(limit)
	}
	if total, ok := paginationMap["total"].(float64); ok {
		pagination.Total = int(total)
	}
	if next, ok := paginationMap["next"].(string); ok {
		pagination.Next = next
	}
	if previous, ok := paginationMap["previous"].(string); ok {
		pagination.Previous = previous
	}

	return pagination
}
