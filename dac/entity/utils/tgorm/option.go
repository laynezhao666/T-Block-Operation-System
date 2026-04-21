package tgorm

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// safeFieldRegexp 用于校验 SQL 字段名是否合法，只允许字母、数字、下划线、点号和反引号。
var safeFieldRegexp = regexp.MustCompile(`^[\w` + "`" + `.]+$`)

// safeField 校验字段名是否安全，防止 SQL 注入。
func safeField(field string) bool {
	return field != "" && safeFieldRegexp.MatchString(field)
}

// invalidFieldOption 当字段名不合法时返回一个带错误的 Option，阻止查询执行。
func invalidFieldOption(field string) Option {
	return func(tx *gorm.DB) *gorm.DB {
		_ = tx.AddError(fmt.Errorf("tgorm: invalid field name %q", field))
		return tx
	}
}

// Option 是一个 GORM scope 函数，用于链式构建查询条件。
type Option func(*gorm.DB) *gorm.DB

// WithOptions 将多个 Option 依次应用到 *gorm.DB 上。
func WithOptions(tx *gorm.DB, opts ...Option) *gorm.DB {
	for _, opt := range opts {
		tx = opt(tx)
	}
	return tx
}

// WithEqual 构造 field = value 条件。
func WithEqual(field string, value interface{}) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(field+" = ?", value)
	}
}

// WithNotEqual 构造 field != value 条件。
func WithNotEqual(field string, value interface{}) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(field+" != ?", value)
	}
}

// WithIn 构造 field IN (values) 条件。
func WithIn(field string, values interface{}) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(field+" IN ?", values)
	}
}

// WithLike 构造 field LIKE %value% 条件。
func WithLike(field string, value string) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(field+" LIKE ?", "%"+value+"%")
	}
}

// WithBetween 构造 field >= begin AND field < end 条件。
func WithBetween(field string, begin, end interface{}) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%v >= ? AND %v < ?", field, field), begin, end)
	}
}

// WithDESC 构造 ORDER BY field DESC。
func WithDESC(field string) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order(field + " desc")
	}
}

// WithASC 构造 ORDER BY field ASC。
func WithASC(field string) Option {
	if !safeField(field) {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order(field + " asc")
	}
}

// WithOr 将多个 Option 用 OR 组合。
// 它会在一个新的 session 上依次应用每个 opt 收集 where 子句，
// 然后将它们合并为一个 OR 条件应用到原始 tx 上。
func WithOr(tx *gorm.DB, opts ...Option) Option {
	return func(db *gorm.DB) *gorm.DB {
		if len(opts) == 0 {
			return db
		}
		// 使用 gorm 的 Or 链式调用构造 OR 条件组
		orQuery := tx.Session(&gorm.Session{NewDB: true})
		for i, opt := range opts {
			sub := opt(tx.Session(&gorm.Session{NewDB: true}))
			if i == 0 {
				orQuery = sub
			} else {
				orQuery = orQuery.Or(sub)
			}
		}
		return db.Where(orQuery)
	}
}

// WithJSONLike 构造 JSON 字段的 LIKE 查询。
// 对于 MySQL JSON 列，使用 JSON_EXTRACT + LIKE 方式匹配。
// columnName: JSON 列名, fields: JSON 路径字段, value: 模糊匹配值。
// 例如 WithJSONLike("channel", ["channel_id"], "abc") 生成:
//
//	JSON_EXTRACT(channel, '$.channel_id') LIKE '%abc%'
func WithJSONLike(columnName string, fields []string, value string) Option {
	if !safeField(columnName) {
		return invalidFieldOption(columnName)
	}
	for _, f := range fields {
		if !safeField(f) {
			return invalidFieldOption(f)
		}
	}
	return func(tx *gorm.DB) *gorm.DB {
		path := "$." + strings.Join(fields, ".")
		return tx.Where(fmt.Sprintf("JSON_EXTRACT(%s, '%s') LIKE ?", columnName, path), "%"+value+"%")
	}
}
