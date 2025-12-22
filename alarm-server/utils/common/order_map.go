package common

// OrderedMap 定义泛型的有序 Map
type OrderedMap[K comparable, V any] struct {
	keys []K     // 保存键的顺序
	data map[K]V // 保存实际的键值对
}

// NewOrderedMap 创建一个新的有序 Map
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		keys: make([]K, 0),
		data: make(map[K]V),
	}
}

// Set 添加或更新键值对
func (om *OrderedMap[K, V]) Set(key K, value V) {
	if _, exists := om.data[key]; !exists {
		om.keys = append(om.keys, key) // 如果是新键，添加到 keys 列表中
	}
	om.data[key] = value
}

// Get 获取键对应的值
func (om *OrderedMap[K, V]) Get(key K) (V, bool) {
	if om == nil {
		return *new(V), false
	}
	value, exists := om.data[key]
	return value, exists
}

// Delete 删除键值对
func (om *OrderedMap[K, V]) Delete(key K) {
	if om == nil {
		return
	}
	delete(om.data, key)
	// 从 keys 列表中移除 key
	for i, k := range om.keys {
		if k == key {
			om.keys = append(om.keys[:i], om.keys[i+1:]...)
			break
		}
	}
}

// Range 遍历有序 Map
func (om *OrderedMap[K, V]) Range(f func(key K, value V)) {
	if om == nil {
		return
	}
	for _, key := range om.keys {
		f(key, om.data[key])
	}
}

// Len 返回 Map 的长度
func (om *OrderedMap[K, V]) Len() int {
	if om == nil {
		return 0
	}
	return len(om.keys)
}
