package common_test

import (
	"testing"

	"agent/logic/collector/device/driver/drivers/common"
)

func TestGetByKeys(t *testing.T) {
	// Sample JSON data structure
	testData := map[string]any{
		"name": "test",
		"data": map[string]any{
			"items": []any{
				map[string]any{"id": "1", "value": 100},
				map[string]any{"id": "2", "value": 200},
			},
			"nested": map[string]any{
				"deep": map[string]any{
					"value": "found",
				},
			},
		},
		"array": []any{
			[]any{"a", "b", "c"},
			[]any{"d", "e", "f"},
		},
		"simple_array": []any{10, 20, 30},
	}

	tests := []struct {
		name     string
		root     any
		keys     []string
		expected any
		wantOk   bool
	}{
		{
			name:     "simple key access",
			root:     testData,
			keys:     []string{"name"},
			expected: "test",
			wantOk:   true,
		},
		{
			name:     "nested key access",
			root:     testData,
			keys:     []string{"data", "nested", "deep", "value"},
			expected: "found",
			wantOk:   true,
		},
		{
			name:     "array index access - numeric",
			root:     testData,
			keys:     []string{"data", "items", "0", "id"},
			expected: "1",
			wantOk:   true,
		},
		{
			name:     "array index access - bracket notation",
			root:     testData,
			keys:     []string{"data", "items[1]", "value"},
			expected: 200,
			wantOk:   true,
		},
		{
			name:     "nested array access",
			root:     testData,
			keys:     []string{"array", "0", "1"},
			expected: "b",
			wantOk:   true,
		},
		{
			name:     "nested array with bracket",
			root:     testData,
			keys:     []string{"array[1][2]"},
			expected: "f",
			wantOk:   true,
		},
		{
			name:     "simple array access",
			root:     testData,
			keys:     []string{"simple_array", "1"},
			expected: 20,
			wantOk:   true,
		},
		{
			name:     "non-existent key",
			root:     testData,
			keys:     []string{"nonexistent"},
			expected: nil,
			wantOk:   false,
		},
		{
			name:     "array index out of bounds",
			root:     testData,
			keys:     []string{"simple_array", "10"},
			expected: nil,
			wantOk:   false,
		},
		{
			name:     "empty key",
			root:     testData,
			keys:     []string{""},
			expected: nil,
			wantOk:   false,
		},
		{
			name:     "whitespace only key",
			root:     testData,
			keys:     []string{"  "},
			expected: nil,
			wantOk:   false,
		},
		{
			name:     "access map key on array",
			root:     testData,
			keys:     []string{"simple_array", "key"},
			expected: nil,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := common.GetByKeys(tt.root, tt.keys)
			if ok != tt.wantOk {
				t.Errorf("GetByKeys() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if tt.wantOk && result != tt.expected {
				t.Errorf("GetByKeys() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseIndexedSegment(t *testing.T) {
	tests := []struct {
		name     string
		seg      string
		wantBase string
		wantIdxs []int
		wantOk   bool
	}{
		{
			name:     "simple key",
			seg:      "name",
			wantBase: "name",
			wantIdxs: nil,
			wantOk:   true,
		},
		{
			name:     "pure number",
			seg:      "0",
			wantBase: "",
			wantIdxs: []int{0},
			wantOk:   true,
		},
		{
			name:     "key with single index",
			seg:      "items[0]",
			wantBase: "items",
			wantIdxs: []int{0},
			wantOk:   true,
		},
		{
			name:     "key with multiple indices",
			seg:      "data[0][1]",
			wantBase: "data",
			wantIdxs: []int{0, 1},
			wantOk:   true,
		},
		{
			name:     "bracket only",
			seg:      "[0]",
			wantBase: "",
			wantIdxs: []int{0},
			wantOk:   true,
		},
		{
			name:     "multiple brackets only",
			seg:      "[1][2][3]",
			wantBase: "",
			wantIdxs: []int{1, 2, 3},
			wantOk:   true,
		},
		{
			name:     "negative index",
			seg:      "[-1]",
			wantBase: "",
			wantIdxs: []int{-1},
			wantOk:   true,
		},
		{
			name:     "large number",
			seg:      "999",
			wantBase: "",
			wantIdxs: []int{999},
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, idxs, ok := common.ParseIndexedSegment(tt.seg)
			if ok != tt.wantOk {
				t.Errorf("ParseIndexedSegment() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if base != tt.wantBase {
				t.Errorf("ParseIndexedSegment() base = %v, want %v", base, tt.wantBase)
			}
			if len(idxs) != len(tt.wantIdxs) {
				t.Errorf("ParseIndexedSegment() idxs length = %v, want %v", len(idxs), len(tt.wantIdxs))
				return
			}
			for i := range idxs {
				if idxs[i] != tt.wantIdxs[i] {
					t.Errorf("ParseIndexedSegment() idxs[%d] = %v, want %v", i, idxs[i], tt.wantIdxs[i])
				}
			}
		})
	}
}

func TestTakeLeadingBracketIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantIdx  int
		wantRest string
		wantErr  bool
	}{
		{
			name:     "simple index",
			input:    "[0]",
			wantIdx:  0,
			wantRest: "",
			wantErr:  false,
		},
		{
			name:     "index with remaining",
			input:    "[5]rest",
			wantIdx:  5,
			wantRest: "rest",
			wantErr:  false,
		},
		{
			name:     "chained indices",
			input:    "[1][2][3]",
			wantIdx:  1,
			wantRest: "[2][3]",
			wantErr:  false,
		},
		{
			name:     "index with spaces",
			input:    "[ 10 ]",
			wantIdx:  10,
			wantRest: "",
			wantErr:  false,
		},
		{
			name:     "no leading bracket",
			input:    "abc",
			wantIdx:  0,
			wantRest: "",
			wantErr:  true,
		},
		{
			name:     "unclosed bracket",
			input:    "[123",
			wantIdx:  0,
			wantRest: "",
			wantErr:  true,
		},
		{
			name:     "empty bracket",
			input:    "[]",
			wantIdx:  0,
			wantRest: "",
			wantErr:  true,
		},
		{
			name:     "non-numeric content",
			input:    "[abc]",
			wantIdx:  0,
			wantRest: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, rest, err := common.TakeLeadingBracketIndex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TakeLeadingBracketIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if idx != tt.wantIdx {
					t.Errorf("TakeLeadingBracketIndex() idx = %v, want %v", idx, tt.wantIdx)
				}
				if rest != tt.wantRest {
					t.Errorf("TakeLeadingBracketIndex() rest = %v, want %v", rest, tt.wantRest)
				}
			}
		})
	}
}

func TestGetByKeysComplexPaths(t *testing.T) {
	// More complex test data mimicking real API responses
	testData := map[string]any{
		"payLoad": map[string]any{
			"data": []any{
				map[string]any{
					"levelValue":    "DC12:02:223325:0101",
					"kw":            123.45,
					"percentageKva": 78.9,
				},
				map[string]any{
					"levelValue":    "DC12:02:223325:0102",
					"kw":            234.56,
					"percentageKva": 89.0,
				},
			},
		},
		"objs": []any{
			map[string]any{
				"guid":       "guid-001",
				"dev_status": 0,
				"tags": []any{
					map[string]any{"tag": "temperature", "value": "25.5", "data_quality": "Good"},
					map[string]any{"tag": "humidity", "value": "60.0", "data_quality": "Good"},
				},
			},
		},
	}

	tests := []struct {
		name     string
		keys     []string
		expected any
		wantOk   bool
	}{
		{
			name:     "payLoad data array first item levelValue",
			keys:     []string{"payLoad", "data", "0", "levelValue"},
			expected: "DC12:02:223325:0101",
			wantOk:   true,
		},
		{
			name:     "payLoad data array second item kw",
			keys:     []string{"payLoad", "data", "1", "kw"},
			expected: 234.56,
			wantOk:   true,
		},
		{
			name:     "objs first item guid",
			keys:     []string{"objs", "0", "guid"},
			expected: "guid-001",
			wantOk:   true,
		},
		{
			name:     "objs tags nested array",
			keys:     []string{"objs", "0", "tags", "1", "tag"},
			expected: "humidity",
			wantOk:   true,
		},
		{
			name:     "objs tags with bracket notation",
			keys:     []string{"objs[0]", "tags[0]", "value"},
			expected: "25.5",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := common.GetByKeys(testData, tt.keys)
			if ok != tt.wantOk {
				t.Errorf("GetByKeys() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if tt.wantOk && result != tt.expected {
				t.Errorf("GetByKeys() = %v, want %v", result, tt.expected)
			}
		})
	}
}
