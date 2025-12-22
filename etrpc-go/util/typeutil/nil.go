// Package typeutil provides various type tools
package typeutil

import "reflect"

// IsNil 判断一个对象是否为空 copy from assert.isNil
func IsNil(object interface{}) bool {

	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	isNilAbleKind := containsKind(
		[]reflect.Kind{
			reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map,
			reflect.Ptr, reflect.Slice},
		kind)

	if isNilAbleKind && value.IsNil() {
		return true
	}

	return false
}

// containsKind copy from assert.isNil
func containsKind(kinds []reflect.Kind, kind reflect.Kind) bool {
	for i := 0; i < len(kinds); i++ {
		if kind == kinds[i] {
			return true
		}
	}

	return false
}
