package expr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr/ast"
)

// transformExpression 转换给定表达式字符串
func transformExpression(expression string) string {
	// 处理 if 语法
	expression = transformIf(expression)

	// 处理逻辑运算符 and/or
	expression = transformLogicalOperators(expression)

	// 处理 avg/min/max 的大小写
	expression = transformFunctions(expression)

	// 处理负变量
	expression = transformNegVar(expression)

	return expression
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

// transformNegVar 将负数变量加上独立括号,如"B+-A"转化为"B+(-A)"
func transformNegVar(exp string) string {
	re := regexp.MustCompile(`([<>]=?)-(\d+)`)
	return re.ReplaceAllStringFunc(exp, func(match string) string {
		return strings.Replace(match, "-", "(-", 1) + ")"
	})
}

// eqRewriter 用于重写函数中的==和!=为eq和neq函数，方便进行1==true这种判断
type eqRewriter struct{}

func (r *eqRewriter) Visit(node *ast.Node) {
	if binary, ok := (*node).(*ast.BinaryNode); ok {
		if binary.Operator == "==" {
			*node = &ast.CallNode{
				Callee:    &ast.IdentifierNode{Value: "eq"},
				Arguments: []ast.Node{binary.Left, binary.Right},
			}
		} else if binary.Operator == "!=" {
			*node = &ast.CallNode{
				Callee:    &ast.IdentifierNode{Value: "neq"},
				Arguments: []ast.Node{binary.Left, binary.Right},
			}
		}
	}
}
