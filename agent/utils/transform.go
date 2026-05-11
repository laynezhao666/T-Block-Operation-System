package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// TransformExpression 转换给定表达式字符串
func TransformExpression(expression string) string {
	// 处理中文标点符号
	expression = transformPunctuation(expression)

	// 处理 if 语法
	expression = transformIf(expression)

	// 处理逻辑运算符 and/or
	expression = transformLogicalOperators(expression)

	// 处理 avg/min/max 的大小写
	expression = transformFunctions(expression)

	return expression
}

// removeSpaces 去除空格
func removeSpaces(expression string) string {
	// 去掉所有多余的空格，但保留表达式中有效的字符间的结构
	return strings.ReplaceAll(expression, " ", "")
}

// transformIf 将if语法转换为三元表达式
func transformIf(expression string) string {
	// 使用正则匹配 if 语法；兼容不规范配置，允许 if 和 ( 之间有空白字符
	re := regexp.MustCompile(`if\s*\(([^,]+),([^,]+),([^)]+)\)`)

	for re.MatchString(expression) {
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			submatches := re.FindStringSubmatch(match)
			condition := strings.TrimSpace(submatches[1])
			trueExpr := strings.TrimSpace(submatches[2])
			falseExpr := strings.TrimSpace(submatches[3])
			return fmt.Sprintf("%s?%s:%s", condition, trueExpr, falseExpr)
		})
	}
	return expression
}

// transformLogicalOperators 替换逻辑运算符 and/or 为 && 和 ||
func transformLogicalOperators(expression string) string {
	// 替换 and 为 &&
	expression = strings.ReplaceAll(expression, " and ", " && ")
	// 替换 or 为 ||
	expression = strings.ReplaceAll(expression, " or ", " || ")
	return expression
}

// transformFunctions 将 Avg/Min/Max 转换为小写
func transformFunctions(expression string) string {
	reAvg := regexp.MustCompile(`\bAvg\b`)
	reMin := regexp.MustCompile(`\bMin\b`)
	reMax := regexp.MustCompile(`\bMax\b`)

	expression = reAvg.ReplaceAllString(expression, "avg")
	expression = reMin.ReplaceAllString(expression, "min")
	expression = reMax.ReplaceAllString(expression, "max")

	return expression
}

// transformPunctuation 将中文标点符号替换为英文标点符号
func transformPunctuation(expression string) string {
	expression = strings.ReplaceAll(expression, "，", ",") // 中文逗号 -> 英文逗号
	expression = strings.ReplaceAll(expression, "（", "(") // 中文左括号 -> 英文左括号
	expression = strings.ReplaceAll(expression, "）", ")") // 中文右括号 -> 英文右括号
	return expression
}
