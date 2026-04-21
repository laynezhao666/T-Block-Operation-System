// Package set 提供集合数据结构的实现。
package set

// StringSet 字符串集合接口，支持增删查和集合运算
type StringSet interface {
	// Contain 检查key是否存在于集合中
	Contain(key string) bool
	// AddWithCheck 添加key到集合，返回是否为新增
	AddWithCheck(key string) bool
	// Add 添加key到集合
	Add(key string)
	// AddSlice 批量添加key到集合
	AddSlice(keys []string)
	// Union 返回与t的并集（新集合）
	Union(s StringSet) StringSet
	// UnionInplace 原地合并t到当前集合
	UnionInplace(s StringSet)
	// DeleteWithCheck 删除key，返回是否存在
	DeleteWithCheck(key string) bool
	// Delete 删除key
	Delete(key string)
	// Get 返回集合中所有元素
	Get() []string
	// Size 返回集合元素个数
	Size() int
	// Shrink 释放多余内存
	Shrink()
	// Clone 深拷贝集合
	Clone() StringSet
	// Foreach 遍历集合中的每个元素
	Foreach(f func(string) error) error
}

// stringSet StringSet接口的map实现
type stringSet struct {
	m map[string]struct{}
}

// NewStringSet 创建空的字符串集合
func NewStringSet() StringSet {
	return &stringSet{
		m: make(map[string]struct{}),
	}
}

// NewStringSetWithData 使用初始数据创建字符串集合
func NewStringSetWithData(d []string) StringSet {
	s := NewStringSet()
	for _, e := range d {
		s.Add(e)
	}
	return s
}

// Contain 检查key是否存在于集合中
func (s *stringSet) Contain(key string) bool {
	if s == nil {
		return false
	}
	_, ok := s.m[key]
	return ok
}

// AddWithCheck 添加key到集合，返回是否为新增
func (s *stringSet) AddWithCheck(key string) bool {
	if s == nil {
		return false
	}
	_, ok := s.m[key]
	if ok {
		return false
	}

	s.m[key] = struct{}{}
	return true
}

// Add 添加key到集合
func (s *stringSet) Add(key string) {
	if s == nil {
		return
	}
	s.m[key] = struct{}{}
}

// AddSlice 批量添加key到集合
func (s *stringSet) AddSlice(keys []string) {
	if s == nil {
		s.m = make(map[string]struct{}, len(keys))
	}
	for _, key := range keys {
		s.m[key] = struct{}{}
	}
}

// Union 返回与t的并集（新集合）
func (s *stringSet) Union(t StringSet) StringSet {
	if s == nil {
		return t.Clone()
	}
	if t == nil {
		return s.Clone()
	}

	u := s.Clone()
	u.UnionInplace(t)
	return u
}

// UnionInplace 原地合并t到当前集合
func (s *stringSet) UnionInplace(t StringSet) {
	if s == nil {
		return
	}
	if s.m == nil {
		s.m = make(map[string]struct{}, t.Size())
	}
	_ = t.Foreach(func(v string) error {
		s.m[v] = struct{}{}
		return nil
	})
}

// DeleteWithCheck 删除key，返回是否存在
func (s *stringSet) DeleteWithCheck(key string) bool {
	if s == nil {
		return false
	}
	if _, ok := s.m[key]; !ok {
		return false
	}

	delete(s.m, key)
	return true
}

// Delete 删除key
func (s *stringSet) Delete(key string) {
	if s == nil {
		return
	}
	delete(s.m, key)
}

// Get 返回集合中所有元素
func (s *stringSet) Get() []string {
	if s == nil {
		return nil
	}
	r := make([]string, 0, len(s.m))
	for k := range s.m {
		r = append(r, k)
	}
	return r
}

// Size 返回集合元素个数
func (s *stringSet) Size() int {
	if s == nil {
		return 0
	}
	return len(s.m)
}

// Shrink 释放多余内存
func (s *stringSet) Shrink() {
	if s == nil {
		return
	}
	m := make(map[string]struct{}, len(s.m))
	for k := range s.m {
		m[k] = struct{}{}
	}
	s.m = m
}

// Clone 深拷贝集合
func (s *stringSet) Clone() StringSet {
	if s == nil {
		return nil
	}

	t := new(stringSet)
	t.m = make(map[string]struct{}, len(s.m))
	for k := range s.m {
		t.m[k] = struct{}{}
	}

	return t
}

// Foreach 遍历集合中的每个元素
func (s *stringSet) Foreach(f func(string) error) error {
	if s == nil {
		return nil
	}
	var err error
	for k := range s.m {
		if err = f(k); err != nil {
			return err
		}
	}
	return nil
}
