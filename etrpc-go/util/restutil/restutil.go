// Package restutil provides restful http method tools
package restutil

// Default judge and return the value
func Default(val, min, max int64) int64 {

	if val <= min {
		return min
	}
	if val >= max {
		return max
	}

	return val
}
