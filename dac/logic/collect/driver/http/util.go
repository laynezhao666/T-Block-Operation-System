package http

import (
	"strings"
	"time"
)

const (
	fetchRecordDuration = time.Hour * 24
	// 最新只取 10s 前的数据
	fetchRecordLast = time.Second * 10
)

// FormatTime 转换为 uri 格式，空格换为 %20
func FormatTime(t time.Time) string {
	return strings.ReplaceAll(t.Format("2006-01-02 15:04:05"), " ", "%20")
}

// getFetchTime 避免门禁时间未同步导致未获取到数据，预期门禁时间偏差在 fetchRecordLast 内
func getFetchTime(t time.Time) time.Time {
	return t.Add(-fetchRecordLast)
}
