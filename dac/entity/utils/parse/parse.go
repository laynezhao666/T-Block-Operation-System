// Package parse 提供基于反射的通用数据结构转换工具。
package parse

import (
	"errors"
	"fmt"
	"reflect"
)

// maxRecursionLevel 最大递归深度限制
const (
	maxRecursionLevel = 1000
)

// errNil 空指针错误
// errMaxRecursion 超过最大递归深度错误
var (
	errNil          = errors.New("nil pointer")
	errMaxRecursion = errors.New("exceed max recursion level")
)

// Parse 将src数据结构解析转换到dst中，支持struct/map/slice等类型
func (p *parser) Parse(dst, src interface{}) error {
	if p == nil {
		return nil
	}
	if dst == nil {
		return nil
	}
	srcValue := reflect.ValueOf(src)
	if !srcValue.IsValid() {
		return nil
	}
	for srcValue.Kind() == reflect.Pointer {
		if srcValue.IsNil() {
			return errNil
		}
		srcValue = reflect.ValueOf(srcValue.Elem().Interface())
	}

	dstPointer := reflect.ValueOf(dst)
	return p.recurse(&dstPointer, &srcValue, nil, nil, 0)
}

// recurse 递归处理数据结构转换
func (p *parser) recurse(dstPointer, srcValue *reflect.Value,
	dstName2Addr, srcName2Value map[string]reflect.Value,
	recurseLevel int,
) (err error) {
	if dstPointer.Kind() != reflect.Pointer {
		return fmt.Errorf("dst is not pointer kind: %v", dstPointer.Kind())
	}

	if recurseLevel > maxRecursionLevel {
		return errMaxRecursion
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	dstValue := dstPointer.Elem()
	dstType := dstValue.Type()
	dstKind := dstValue.Kind()

	srcType := srcValue.Type()
	srcKind := srcValue.Kind()

	// 解引用interface类型的源值
	if dstKind != reflect.Interface {
		for srcKind == reflect.Interface {
			var newSrc reflect.Value
			if srcValue.IsNil() {
				newSrc = reflect.New(dstType).Elem()
			} else {
				newSrc = reflect.ValueOf(srcValue.Elem().Interface())
			}
			srcValue = &newSrc
			srcType = newSrc.Type()
			srcKind = newSrc.Kind()
		}
	}

	// 根据目标类型分发到对应的转换方法
	switch dstKind {
	case reflect.Slice:
		// 切片类型转换
		switch srcKind {
		case reflect.Slice, reflect.Array:
			return p.slice2Slice(&dstValue, srcValue, dstType, recurseLevel)
		default:
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
	case reflect.Array:
		// 数组类型转换
		switch srcKind {
		case reflect.Slice, reflect.Array:
			return p.slice2Array(&dstValue, srcValue, recurseLevel)
		default:
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
	case reflect.Struct:
		// 结构体类型转换（支持map→struct和struct→struct）
		switch srcKind {
		case reflect.Map:
			return p.map2Struct(&dstValue, srcValue, dstType, recurseLevel)
		case reflect.Struct:
			return p.struct2Struct(&dstValue, srcValue, dstType, srcType, &dstName2Addr, &srcName2Value, recurseLevel)
		default:
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
	case reflect.Map:
		// Map类型转换
		if srcKind != reflect.Map {
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
		return p.map2Map(&dstValue, srcValue, dstType, recurseLevel)
	default:
		// 基本类型：直接赋值或类型转换
		if p.config.DisableConvert {
			dstValue.Set(*srcValue)
		} else {
			if srcValue.CanConvert(dstType) {
				dstValue.Set(srcValue.Convert(dstType))
			}
		}
	}

	return err
}

// map2Struct 将map转换为struct
func (p *parser) map2Struct(dstValue, srcValue *reflect.Value,
	dstType reflect.Type, recurseLevel int,
) error {
	var err error
	dstFieldNum := dstType.NumField()
	for i := 0; i < dstFieldNum; i++ {
		dstFieldType := dstType.Field(i)
		if !dstFieldType.IsExported() {
			continue
		}
		if dstFieldType.Anonymous {
			dstFieldValueAddr := dstValue.Field(i).Addr()
			if err = p.recurse(&dstFieldValueAddr, srcValue, nil, nil, recurseLevel+1); err != nil {
				return err
			}
		} else {
			dstFieldName, ok := p.config.FieldExtractor(&dstFieldType)
			if !ok {
				continue
			}
			newSrcValue := srcValue.MapIndex(reflect.ValueOf(dstFieldName))
			if !newSrcValue.IsValid() {
				continue
			}
			if !newSrcValue.CanInterface() {
				continue
			}
			dstFieldValueAddr := dstValue.Field(i).Addr()
			if err = p.recurse(&dstFieldValueAddr, &newSrcValue, nil, nil, recurseLevel+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// map2Map 将map转换为另一种map类型
func (p *parser) map2Map(dstValue, srcValue *reflect.Value,
	dstType reflect.Type, recurseLevel int,
) error {
	var err error
	l := srcValue.Len()
	dstValue.Set(reflect.MakeMapWithSize(dstType, l))
	iter := srcValue.MapRange()
	for iter.Next() {
		value := iter.Value()
		newValue := reflect.New(dstType.Elem()).Elem()
		newValueAddr := newValue.Addr()
		if err = p.recurse(&newValueAddr, &value, nil, nil, recurseLevel+1); err != nil {
			return err
		}
		dstValue.SetMapIndex(iter.Key(), newValue)
	}
	return nil
}

// slice2Slice 将slice转换为另一种slice类型
func (p *parser) slice2Slice(dstValue, srcValue *reflect.Value,
	dstType reflect.Type, recurseLevel int,
) error {
	var err error
	l := srcValue.Len()
	dstValue.Set(reflect.MakeSlice(dstType, l, l))
	for i := 0; i < l; i++ {
		e := dstValue.Index(i).Addr()
		s := srcValue.Index(i)
		if err = p.recurse(&e, &s, nil, nil, recurseLevel+1); err != nil {
			return err
		}
	}
	return nil
}

// slice2Array 将slice转换为array类型
func (p *parser) slice2Array(dstValue, srcValue *reflect.Value,
	recurseLevel int,
) error {
	var err error
	srcLen := srcValue.Len()
	minLen := dstValue.Len()
	if minLen > srcLen {
		minLen = srcLen
	}
	for i := minLen - 1; i >= 0; i++ {
		e := dstValue.Index(i).Addr()
		s := srcValue.Index(i)
		if err = p.recurse(&e, &s, nil, nil, recurseLevel+1); err != nil {
			return err
		}
	}
	return nil
}

// struct2Struct 将struct转换为另一种struct类型
func (p *parser) struct2Struct(dstValue, srcValue *reflect.Value,
	dstType, srcType reflect.Type,
	dstName2Addr, srcName2Value *map[string]reflect.Value,
	recurseLevel int,
) error {
	var err error
	if *dstName2Addr == nil {
		if *dstName2Addr, err = p.getDstStructFieldNameAddr(dstValue, dstType); err != nil {
			return err
		}
	}
	name2Addr := *dstName2Addr

	if *srcName2Value == nil {
		if *srcName2Value, err = p.getSrcStructFieldNameValue(srcValue, srcType, name2Addr); err != nil {
			return err
		}
	}
	name2Value := *srcName2Value

	for name, value := range name2Value {
		addr := name2Addr[name]
		if err = p.recurse(&addr, &value, nil, nil, recurseLevel+1); err != nil {
			return err
		}
	}

	return nil
}
