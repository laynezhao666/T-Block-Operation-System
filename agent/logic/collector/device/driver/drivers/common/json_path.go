// Package common provides reusable utilities for HTTP-based drivers
package common

import (
	"fmt"
	"strconv"
	"strings"
)

// GetByKeys supports multi-level JSON path parsing
// Features:
// - map key access: a.b.c
// - array index access: data.0.value
// - bracket notation: data[0].value or items[1][2].name
// Parameters:
// - root: JSON root node
// - keys: path keys array
// Returns:
// - any: parsed value
// - bool: whether successfully found
func GetByKeys(root any, keys []string) (any, bool) {
	cur := root
	for _, seg := range keys {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			return nil, false
		}

		// Allow segments like "data[0][1]" with multiple indices
		base, idxs, ok := ParseIndexedSegment(seg)
		if !ok {
			return nil, false
		}

		// First access map with base (base may be empty for direct array index like "[0]")
		if base != "" {
			m, mok := cur.(map[string]any)
			if !mok {
				return nil, false
			}
			v, ex := m[base]
			if !ex {
				return nil, false
			}
			cur = v
		}

		// Apply all indices sequentially
		for _, idx := range idxs {
			arr, aok := cur.([]any)
			if !aok || idx < 0 || idx >= len(arr) {
				return nil, false
			}
			cur = arr[idx]
		}
	}

	return cur, true
}

// ParseIndexedSegment parses path segment with indices
// Supports:
// - "key", "key[0]", "key[0][1]", "[0]", "0" (pure number)
// Parameters:
// - seg: path segment string
// Returns:
// - base: base key name (may be empty)
// - idxs: index chain
// - ok: whether parsing succeeded
func ParseIndexedSegment(seg string) (base string, idxs []int, ok bool) {
	s := seg

	// Pure number: treat as direct array index (equivalent to base="" + idx)
	if i, e := strconv.Atoi(s); e == nil {
		return "", []int{i}, true
	}

	// Support [0] prefix form (equivalent to base="")
	if strings.HasPrefix(s, "[") {
		base = ""
		var rest = s
		for len(rest) > 0 {
			i, r, e := TakeLeadingBracketIndex(rest)
			if e != nil {
				return "", nil, false
			}
			idxs = append(idxs, i)
			rest = r
			if rest == "" {
				break
			}
		}
		return base, idxs, true
	}

	// Regular "key" or "key[0][1]" form
	// First extract base (up to first '[')
	if p := strings.IndexByte(s, '['); p >= 0 {
		base = s[:p]
		rest := s[p:]
		for len(rest) > 0 {
			i, r, e := TakeLeadingBracketIndex(rest)
			if e != nil {
				return "", nil, false
			}
			idxs = append(idxs, i)
			rest = r
			if rest == "" {
				break
			}
		}
		return base, idxs, true
	}

	// Simple "key"
	return s, nil, true
}

// TakeLeadingBracketIndex extracts leading "[number]"
// Parameters:
// - s: input string
// Returns:
// - idx: index value
// - rest: remaining string
// - err: parse error
func TakeLeadingBracketIndex(s string) (idx int, rest string, err error) {
	if !strings.HasPrefix(s, "[") {
		return 0, "", fmt.Errorf("no leading '['")
	}
	end := strings.IndexByte(s, ']')
	if end <= 1 {
		return 0, "", fmt.Errorf("unclosed or empty bracket")
	}
	num := s[1:end]
	i, parseErr := strconv.Atoi(strings.TrimSpace(num))
	if parseErr != nil {
		return 0, "", parseErr
	}
	return i, s[end+1:], nil
}
