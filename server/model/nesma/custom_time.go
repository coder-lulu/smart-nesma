package nesma

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// CustomTime 自定义时间类型，支持多种格式解析
type CustomTime struct {
	time.Time
}

// UnmarshalJSON 自定义JSON反序列化，支持多种时间格式
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// 移除引号
	str := strings.Trim(string(data), `"`)
	if str == "" {
		return nil
	}

	// 支持的时间格式
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 without timezone
		"2006-01-02T15:04:05",       // ISO 8601 without timezone
		"2006-01-02 15:04:05",       // MySQL datetime format
		"2006-01-02",                // Date only
		time.RFC3339,                // RFC3339
		time.RFC3339Nano,            // RFC3339Nano
	}

	var err error
	for _, format := range formats {
		ct.Time, err = time.Parse(format, str)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("无法解析时间格式: %s", str)
}

// MarshalJSON 自定义JSON序列化
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ct.Time.Format(time.RFC3339) + `"`), nil
}

// Value 实现driver.Valuer接口，用于数据库存储
func (ct CustomTime) Value() (driver.Value, error) {
	if ct.Time.IsZero() {
		return nil, nil
	}
	return ct.Time, nil
}

// Scan 实现sql.Scanner接口，用于数据库读取
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		return nil
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		ct.Time = t
		return nil
	default:
		return fmt.Errorf("无法将 %T 转换为 CustomTime", value)
	}
}
