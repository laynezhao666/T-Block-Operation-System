package utils

// SetIfNotNull 设置指针的值
func SetIfNotNull(p *bool, v bool) {
	if p != nil {
		*p = v
	}
}
