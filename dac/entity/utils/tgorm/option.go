package tgorm

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// safeFieldRegexp 用于校验纯字段名是否合法，仅允许字母、数字和下划线。
var safeFieldRegexp = regexp.MustCompile(`^\w+$`)

// safeFieldPathRegexp 用于校验带点号的字段路径（如 table.column），每段仅允许字母、数字和下划线。
var safeFieldPathRegexp = regexp.MustCompile(`^\w+(\.\w+)*$`)

// safeField 校验纯字段名是否安全（白名单），防止 SQL 注入。
func safeField(field string) bool {
	return field != "" && safeFieldRegexp.MatchString(field)
}

// safeFieldPath 校验带点号的字段路径是否安全（白名单）。
func safeFieldPath(field string) bool {
	return field != "" && safeFieldPathRegexp.MatchString(field)
}

// sanitizeField 对外部传入的字段名做白名单校验，并转换成 clause.Column 结构体。
//  1. 通过严格正则白名单拦截任何含特殊字符的输入；
//  2. 校验通过后将字符串转为结构体字段（Table + Name），
//     切断“用户输入字符串 -> SQL 文本”的直接数据流；
//  3. 后续所有 Where 条件均以 clause.* 结构体形式传入 GORM，
//     不再发生字符串拼接，从根源消除 SQL 注入风险。
func sanitizeField(field string) (clause.Column, bool) {
	if !safeFieldPath(field) {
		return clause.Column{}, false
	}
	// 支持 table.column 形式，分离 Table 与 Name。
	if idx := strings.Index(field, "."); idx > 0 {
		return clause.Column{Table: field[:idx], Name: field[idx+1:]}, true
	}
	return clause.Column{Name: field}, true
}

// invalidFieldOption 当字段名不合法时返回带错误的 Option，阻止 SQL 执行。
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
// 使用 clause.Eq 结构体传参，无字符串拼接。
func WithEqual(field string, value interface{}) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(clause.Eq{Column: col, Value: value})
	}
}

// WithNotEqual 构造 field != value 条件。
// 使用 clause.Neq 结构体传参，无字符串拼接。
func WithNotEqual(field string, value interface{}) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(clause.Neq{Column: col, Value: value})
	}
}

// WithIn 构造 field IN (values) 条件。
//
// 实现说明（安全性）：
//  1. 字段名经过 sanitizeField 白名单校验并转换成 clause.Column 结构体，
//     不会以字符串形式参与 SQL 拼接；
//  2. values 通过反射归一化为 []interface{} 后，交由 GORM 原生的
//     clause.IN 结构体（Column + Values）传参，SQL 片段由 GORM 的
//     clause builder 完全托管生成，调用方无法注入任何 SQL 文本；
//  3. 相比于基于 gorm.Expr("? IN (?)", ...) 的写法，本实现彻底消除了
//     字符串 SQL 模板，静态扫描工具（如 pecker-go）的污点数据流追踪
//     会在 clause.IN 结构体字段赋值处被切断。
func WithIn(field string, values interface{}) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	// 将任意类型的 values 归一化为 []interface{}，以适配 clause.IN。
	normalized := normalizeInValues(values)
	return func(tx *gorm.DB) *gorm.DB {
		// 空集合场景下，IN () 在 MySQL 语义上恒为 false，
		// 这里保持与原实现一致：依然交给 GORM 处理（避免语义差异）。
		return tx.Where(clause.IN{Column: col, Values: normalized})
	}
}

// normalizeInValues 将任意类型 values 转换为 []interface{}。
//   - 若为 nil，返回 nil；
//   - 若为 slice/array，按元素展开；
//   - 否则当作单值包装。
//
// 该函数仅做类型归一化，不触碰 SQL 文本，安全。
func normalizeInValues(values interface{}) []interface{} {
	if values == nil {
		return nil
	}
	rv := reflect.ValueOf(values)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		length := rv.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result
	default:
		return []interface{}{values}
	}
}

// WithLike 构造 field LIKE %value% 条件。
// 使用 clause.Like 结构体传参，无字符串拼接。
func WithLike(field string, value string) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(clause.Like{Column: col, Value: "%" + value + "%"})
	}
}

// WithBetween 构造 field >= begin AND field < end 条件。
// 使用 clause.Gte 与 clause.Lt 两个结构体组合，无字符串拼接。
func WithBetween(field string, begin, end interface{}) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(clause.Gte{Column: col, Value: begin}).
			Where(clause.Lt{Column: col, Value: end})
	}
}

// WithDESC 构造 ORDER BY field DESC。
// 使用 clause.OrderBy 结构体传参，无字符串拼接。
func WithDESC(field string) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Clauses(clause.OrderBy{
			Columns: []clause.OrderByColumn{{Column: col, Desc: true}},
		})
	}
}

// WithASC 构造 ORDER BY field ASC。
// 使用 clause.OrderBy 结构体传参，无字符串拼接。
func WithASC(field string) Option {
	col, ok := sanitizeField(field)
	if !ok {
		return invalidFieldOption(field)
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Clauses(clause.OrderBy{
			Columns: []clause.OrderByColumn{{Column: col, Desc: false}},
		})
	}
}

// WithOr 将多个 Option 用 OR 组合。
//
// 实现说明（安全性）：
//  1. 每个子 Option 在独立 session（NewDB: true）上执行，产出的
//     *gorm.DB 仅作为“条件收集容器”——随后只从其 Statement.Clauses
//     中读取结构化的 clause.Expression，不再把 *gorm.DB 整体回塞
//     给主查询的 Where；
//  2. 最终通过 clause.OrConditions / clause.AndConditions 结构体
//     组合，并使用 db.Clauses(clause.Where{...}) 应用到主查询，
//     避免了“*gorm.DB -> Where(*gorm.DB)”这一被静态扫描器标记为
//     污点传播的路径，彻底消除 SQL Injection 的误报与潜在风险；
//  3. 字段名、参数值等底层仍由各子 Option 通过 clause.* 结构体承载，
//     整条链路不存在任何用户字符串拼接进入 SQL 文本的环节。
func WithOr(tx *gorm.DB, opts ...Option) Option {
	return func(db *gorm.DB) *gorm.DB {
		if len(opts) == 0 {
			return db
		}
		// 收集每个子 Option 产生的 WHERE 表达式组，每组用 AND 内部连接，
		// 整体再用 OR 连接。
		orExprs := collectOrExpressions(tx, opts)
		if len(orExprs) == 0 {
			return db
		}
		if len(orExprs) == 1 {
			// 只有一组条件，等价于普通 AND Where，直接应用即可。
			return db.Clauses(clause.Where{Exprs: orExprs})
		}
		// 多组条件，整体包裹为一个 OrConditions 节点再作为单个
		// WHERE 表达式下发，保证与原 OR 语义一致。
		return db.Clauses(clause.Where{
			Exprs: []clause.Expression{clause.OrConditions{Exprs: orExprs}},
		})
	}
}

// collectOrExpressions 在独立 session 上执行每个子 Option，
// 并从其 Statement.Clauses["WHERE"] 中提取 clause.Expression 列表。
// 多条件（AND 组）会被包裹为 clause.AndConditions，保证子组原子性。
func collectOrExpressions(tx *gorm.DB, opts []Option) []clause.Expression {
	const whereName = "WHERE"
	result := make([]clause.Expression, 0, len(opts))
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		sub := opt(tx.Session(&gorm.Session{NewDB: true}))
		if sub == nil || sub.Statement == nil {
			continue
		}
		whereClause, ok := sub.Statement.Clauses[whereName]
		if !ok {
			continue
		}
		where, ok := whereClause.Expression.(clause.Where)
		if !ok || len(where.Exprs) == 0 {
			continue
		}
		if len(where.Exprs) == 1 {
			result = append(result, where.Exprs[0])
			continue
		}
		// 多条件在子 Option 内部是 AND 连接，包裹后作为一个整体参与 OR。
		result = append(result, clause.AndConditions{Exprs: where.Exprs})
	}
	return result
}

// WithJSONLike 构造 JSON 字段的 LIKE 查询。
// 对于 MySQL JSON 列，使用 JSON_EXTRACT + LIKE 方式匹配。
//
//	columnName: JSON 列名（须为合法标识符）
//	fields:     JSON 路径字段（每段须为合法标识符）
//	value:      模糊匹配值（走参数化绑定）
//
// 例如 WithJSONLike("channel", ["channel_id"], "abc") 生成:
//
//	JSON_EXTRACT(`channel`, '$.channel_id') LIKE '%abc%'
//
// 实现上通过 gorm.Expr 将 clause.Column 和路径、匹配值全部作为参数占位符注入，
// SQL 模板为纯静态字面量，不存在任何用户输入的字符串拼接。
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
		// 各字段名均已通过严格白名单（^\w+$）校验，
		// path 由校验过的静态片段拼接，仅包含 [A-Za-z0-9_.] 字符。
		path := "$." + strings.Join(fields, ".")
		return tx.Where(gorm.Expr(
			"JSON_EXTRACT(?, ?) LIKE ?",
			clause.Column{Name: columnName}, path, "%"+value+"%",
		))
	}
}
