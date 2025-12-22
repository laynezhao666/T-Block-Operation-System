package generator

// StaticImpl 静态数据生成器
type StaticImpl struct {
	x float64
}

// NewStaticImpl 创建静态数据生成器
func NewStaticImpl(x float64) *StaticImpl {
	return &StaticImpl{x: x}
}

// Get 获取数据
func (s *StaticImpl) Get() float64 {
	return s.x
}
