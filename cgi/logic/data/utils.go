package data

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

const (
	PointKeyConditionName          = "point_key"
	PointIdConditionName           = "point_id"
	DeviceGidConditionName         = "device_gid"
	PointIdEnConditionName 		   = "point_id_en"
	DeviceTypeZhConditionName      = "device_type_zh"
	ApplicationTypeZhConditionName = "application_type_zh"
	// GidPointConnectionChar gid和测点类型连接符
	GidPointConnectionChar = "."
	DefaultPointValue      = "--"
	// TimeFormat 统一时间格式
	TimeFormat = "2006-01-02 15:04:05"
	// LocalLocation 本地时区
	LocalLocation = "Asia/Shanghai"
)

func stringToDate(dateString string) (int64, error) {
	// 加载指定的时区
	location, err := time.LoadLocation(LocalLocation)
	if err != nil {
		return 0, fmt.Errorf("failed to load location: %v", err)
	}

	// 解析时间字符串，使用指定的时区
	date, err := time.ParseInLocation(TimeFormat, dateString, location)
	if err != nil {
		return 0, fmt.Errorf("failed to parse date string: %v", err)
	}

	// 将时间转换为 int64 类型的时间戳
	timestamp := date.Unix()
	return timestamp, nil
}

func dataToString(timestamp int64) (string, error) {
	// 将 int64 类型的时间戳转换为 time.Time 类型
	date := time.Unix(timestamp, 0)
	// 将 time.Time 类型转换为 string 类型
	dateString := date.Format(TimeFormat)
	return dateString, nil
}

func pointKeyToGid(pointKey string) (string, error) {
	list := strings.Split(pointKey, GidPointConnectionChar)
	if len(list) != 2 {
		return "", fmt.Errorf("point key format err,pointKey: %s", pointKey)
	}
	return list[0], nil
}

func isAllChinese(str string) bool {
	for _, r := range str {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}

// InArray 是否在数组中
func InArray(str string, arr []string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}