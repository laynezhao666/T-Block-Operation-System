package common_test

import (
	"testing"
	"time"

	"agent/logic/collector/device/driver/drivers/common"
)

func TestBaseCacheItem_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		expireTime time.Time
		wantResult bool
	}{
		{
			name:       "not expired - future time",
			expireTime: time.Now().Add(1 * time.Hour),
			wantResult: false,
		},
		{
			name:       "expired - past time",
			expireTime: time.Now().Add(-1 * time.Hour),
			wantResult: true,
		},
		{
			name:       "just expired - 1 second ago",
			expireTime: time.Now().Add(-1 * time.Second),
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &common.BaseCacheItem{
				Value:      "test",
				ExpireTime: tt.expireTime,
			}
			if got := item.IsExpired(); got != tt.wantResult {
				t.Errorf("BaseCacheItem.IsExpired() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestCacheState_Operations(t *testing.T) {
	t.Run("IncrementFailure", func(t *testing.T) {
		state := &common.CacheState{}
		if state.FailureCount != 0 {
			t.Errorf("Initial FailureCount should be 0, got %v", state.FailureCount)
		}

		state.IncrementFailure()
		if state.FailureCount != 1 {
			t.Errorf("FailureCount after increment should be 1, got %v", state.FailureCount)
		}

		state.IncrementFailure()
		state.IncrementFailure()
		if state.FailureCount != 3 {
			t.Errorf("FailureCount after 3 increments should be 3, got %v", state.FailureCount)
		}
	})

	t.Run("ResetFailure", func(t *testing.T) {
		state := &common.CacheState{FailureCount: 5}
		beforeReset := time.Now()
		state.ResetFailure()

		if state.FailureCount != 0 {
			t.Errorf("FailureCount after reset should be 0, got %v", state.FailureCount)
		}
		if state.LastUpdate.Before(beforeReset) {
			t.Errorf("LastUpdate should be after reset time")
		}
	})

	t.Run("IsHealthy", func(t *testing.T) {
		tests := []struct {
			name         string
			failureCount int
			maxFailures  int
			wantHealthy  bool
		}{
			{
				name:         "healthy - no failures",
				failureCount: 0,
				maxFailures:  3,
				wantHealthy:  true,
			},
			{
				name:         "healthy - below threshold",
				failureCount: 2,
				maxFailures:  3,
				wantHealthy:  true,
			},
			{
				name:         "unhealthy - at threshold",
				failureCount: 3,
				maxFailures:  3,
				wantHealthy:  false,
			},
			{
				name:         "unhealthy - above threshold",
				failureCount: 5,
				maxFailures:  3,
				wantHealthy:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				state := &common.CacheState{FailureCount: tt.failureCount}
				if got := state.IsHealthy(tt.maxFailures); got != tt.wantHealthy {
					t.Errorf("CacheState.IsHealthy() = %v, want %v", got, tt.wantHealthy)
				}
			})
		}
	})
}

func TestParsePaginationFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		respData map[string]interface{}
		wantNil  bool
		expected *common.PaginationInfo
	}{
		{
			name:     "no pagination field",
			respData: map[string]interface{}{"data": "test"},
			wantNil:  true,
		},
		{
			name: "pagination not a map",
			respData: map[string]interface{}{
				"pagination": "invalid",
			},
			wantNil: true,
		},
		{
			name: "valid full pagination",
			respData: map[string]interface{}{
				"pagination": map[string]interface{}{
					"offset":   float64(0),
					"limit":    float64(100),
					"total":    float64(500),
					"next":     "/api/data?offset=100",
					"previous": "",
				},
			},
			wantNil: false,
			expected: &common.PaginationInfo{
				Offset:   0,
				Limit:    100,
				Total:    500,
				Next:     "/api/data?offset=100",
				Previous: "",
			},
		},
		{
			name: "partial pagination - offset and limit only",
			respData: map[string]interface{}{
				"pagination": map[string]interface{}{
					"offset": float64(100),
					"limit":  float64(50),
				},
			},
			wantNil: false,
			expected: &common.PaginationInfo{
				Offset: 100,
				Limit:  50,
			},
		},
		{
			name: "large pagination values",
			respData: map[string]interface{}{
				"pagination": map[string]interface{}{
					"offset": float64(10000),
					"limit":  float64(1000),
					"total":  float64(50000),
				},
			},
			wantNil: false,
			expected: &common.PaginationInfo{
				Offset: 10000,
				Limit:  1000,
				Total:  50000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.ParsePaginationFromResponse(tt.respData)
			if tt.wantNil {
				if result != nil {
					t.Errorf("ParsePaginationFromResponse() = %v, want nil", result)
				}
				return
			}
			if result == nil {
				t.Errorf("ParsePaginationFromResponse() = nil, want non-nil")
				return
			}
			if result.Offset != tt.expected.Offset {
				t.Errorf("Offset = %v, want %v", result.Offset, tt.expected.Offset)
			}
			if result.Limit != tt.expected.Limit {
				t.Errorf("Limit = %v, want %v", result.Limit, tt.expected.Limit)
			}
			if result.Total != tt.expected.Total {
				t.Errorf("Total = %v, want %v", result.Total, tt.expected.Total)
			}
			if result.Next != tt.expected.Next {
				t.Errorf("Next = %v, want %v", result.Next, tt.expected.Next)
			}
			if result.Previous != tt.expected.Previous {
				t.Errorf("Previous = %v, want %v", result.Previous, tt.expected.Previous)
			}
		})
	}
}
