// Package generator 生成器
package generator

type impl interface {
	Get() float64
}

// Generator 生成器
type Generator struct {
	currentValue float64
	impl         impl
}

// New 创建生成器
func New(impl impl) *Generator {
	return &Generator{
		currentValue: 0,
		impl:         impl,
	}
}

// GetCurrentValue 获取当前值
func (g *Generator) GetCurrentValue() float64 {
	return g.currentValue
}

// Generate 生成数据
func (g *Generator) Generate(old *float64) float64 {
	if old != nil {
		*old = g.currentValue
	}
	g.currentValue = g.impl.Get()
	return g.currentValue
}
