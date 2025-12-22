package model

// ConfigVersion 配置版本
type ConfigVersion struct {
	Collector string `json:"collector"`
	Point     string `json:"point"`
}

// Copy 复制
func (c *ConfigVersion) Copy() *ConfigVersion {
	if c == nil {
		return nil
	}
	newCV := &ConfigVersion{
		Collector: c.Collector,
		Point:     c.Point,
	}
	return newCV
}
