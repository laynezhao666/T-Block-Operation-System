package common

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type JsonDate time.Time

func (j JsonDate) String() string {
	t := time.Time(j)
	//unix时间戳为0的时候显示为空
	if t.IsZero() || t.Unix() == 0 {
		return ""
	}

	str := t.Format("2006-01-02 15:04:05")
	if str == "0001-01-01 00:00:00" {
		// 有异常的时间日期，返回空字符串
		return ""
	}
	return str
}

// MarshalJSON 实现json序列化
func (j JsonDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.String() + `"`), nil
}

// UnmarshalJSON 实现json反序列化
func (j *JsonDate) UnmarshalJSON(value []byte) error {
	var v = strings.TrimSpace(strings.Trim(string(value), "\""))
	if v != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			return err
		}
		*j = JsonDate(t)
	}
	return nil
}

// Unix 返回时间戳
func (j *JsonDate) Unix() int64 {
	return (*time.Time)(j).Unix()
}

// UnixOrZero 返回时间戳，如果时间戳为0，则返回0
func (j *JsonDate) UnixOrZero() int64 {
	if t := time.Time(*j); t.IsZero() {
		return 0
	} else {
		return t.Unix()
	}
}

// JsonDBTime json序列化时间类型
type JsonDBTime struct {
	JsonDate
}

// Value 实现driver.Valuer
func (t JsonDBTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if time.Time(t.JsonDate).UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return time.Time(t.JsonDate), nil
}

// Scan 实现driver.Valuer
func (t *JsonDBTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JsonDBTime{JsonDate: JsonDate(time.Time(value))}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
