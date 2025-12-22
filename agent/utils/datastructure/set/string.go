package set

// StringSet StringSet接口
type StringSet interface {
	// Contain 测试 key 是否已存在
	Contain(key string) bool
	// AddWithCheck 添加 key，添加成功返回 true，已存在返回 false
	AddWithCheck(key string) bool
	// Add 添加 key
	Add(key string)
	// AddSlice 添加 slice
	AddSlice(keys []string)
	// Union 求与 s 的并集
	Union(s StringSet) StringSet
	// UnionInplace 原地求与 s 的并集
	UnionInplace(s StringSet)
	// DeleteWithCheck 删除 key，删除成功返回 true，不存在返回 false
	DeleteWithCheck(key string) bool
	// Delete 删除 key
	Delete(key string)
	// Get 返回所有元素
	Get() []string
	// Size 返回元素个数
	Size() int
	// Shrink 释放多余内存
	Shrink()
	// Clone 返回拷贝
	Clone() StringSet
	// Foreach 遍历
	Foreach(f func(string) error) error
}

type stringSet struct {
	m map[string]struct{}
}

// NewStringSet 创建一个 StringSet
func NewStringSet() StringSet {
	return &stringSet{
		m: make(map[string]struct{}),
	}
}

// NewStringSetWithData 创建一个 StringSet
func NewStringSetWithData(d []string) StringSet {
	s := NewStringSet()
	for _, e := range d {
		s.Add(e)
	}
	return s
}

// Contain 测试 key 是否已存在
func (s *stringSet) Contain(key string) bool {
	if s == nil {
		return false
	}
	_, ok := s.m[key]
	return ok
}

// AddWithCheck 添加 key，添加成功返回 true，已存在返回 false
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

// Add 添加 key
func (s *stringSet) Add(key string) {
	if s == nil {
		return
	}
	s.m[key] = struct{}{}
}

// AddSlice 添加 slice
func (s *stringSet) AddSlice(keys []string) {
	if s == nil {
		s.m = make(map[string]struct{}, len(keys))
	}
	for _, key := range keys {
		s.m[key] = struct{}{}
	}
}

// Union 并集
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

// UnionInplace 并集
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

// DeleteWithCheck 删除 key
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

// Delete 删除 key
func (s *stringSet) Delete(key string) {
	if s == nil {
		return
	}
	delete(s.m, key)
}

// Get 获取所有 key
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

// Size 获取大小
func (s *stringSet) Size() int {
	if s == nil {
		return 0
	}
	return len(s.m)
}

// Shrink 缩减内存
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

// Clone 复制
func (s *stringSet) Clone() StringSet {
	if s == nil {
		return nil
	}
	t := NewStringSet()
	_ = s.Foreach(func(k string) error {
		t.Add(k)
		return nil
	})
	return t
}

// Foreach 遍历
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
