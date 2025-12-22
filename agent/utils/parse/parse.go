// Package parse provides json parser
package parse

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	maxRecursionLevel = 1000
)

var (
	errNil          = errors.New("nil pointer")
	errMaxRecursion = errors.New("exceed max recursion level")
)

// Parse 将 src 的值赋给 dst
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

func (p *parser) recurse(dstPointer, srcValue *reflect.Value, dstName2Addr, srcName2Value map[string]reflect.Value,
	recurseLevel int) (err error) {
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

	for (dstKind != reflect.Interface) && srcKind == reflect.Interface {
		newSrc := reflect.ValueOf(srcValue.Elem().Interface())
		srcValue = &newSrc
		srcType = newSrc.Type()
		srcKind = newSrc.Kind()
	}

	switch dstKind {
	case reflect.Slice:
		switch srcKind {
		case reflect.Slice, reflect.Array:
			return p.slice2Slice(&dstValue, srcValue, dstType, recurseLevel)
		default:
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
	case reflect.Array:
		switch srcKind {
		case reflect.Slice, reflect.Array:
			return p.slice2Array(&dstValue, srcValue, recurseLevel)
		default:
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
	case reflect.Struct:
		switch srcKind {
		case reflect.Map:
			return p.map2Struct(&dstValue, srcValue, dstType, recurseLevel)
		case reflect.Struct:
			return p.struct2Struct(&dstValue, srcValue, dstType, srcType, dstName2Addr, srcName2Value, recurseLevel)
		default:
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
	case reflect.Map:
		if srcKind != reflect.Map {
			return fmt.Errorf("src kind: %v, dst kind: %v", srcKind, dstValue.Kind())
		}
		return p.map2Map(&dstValue, srcValue, dstType, recurseLevel)
	default:
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

func (p *parser) map2Struct(dstValue, srcValue *reflect.Value, dstType reflect.Type, recurseLevel int) error {
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
				// map 中不存在对应 key
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

func (p *parser) map2Map(dstValue, srcValue *reflect.Value, dstType reflect.Type, recurseLevel int) error {
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

func (p *parser) slice2Slice(dstValue, srcValue *reflect.Value, dstType reflect.Type, recurseLevel int) error {
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

func (p *parser) slice2Array(dstValue, srcValue *reflect.Value, recurseLevel int) error {
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

func (p *parser) struct2Struct(dstValue, srcValue *reflect.Value, dstType, srcType reflect.Type,
	dstName2Addr, srcName2Value map[string]reflect.Value, recurseLevel int) error {
	var err error
	if dstName2Addr == nil {
		if dstName2Addr, err = p.getDstStructFieldNameAddr(dstValue, dstType); err != nil {
			return err
		}
	}

	if srcName2Value == nil {
		if srcName2Value, err = p.getSrcStructFieldNameValue(srcValue, srcType, dstName2Addr); err != nil {
			return err
		}
	}

	for name, value := range srcName2Value {
		addr := dstName2Addr[name]
		if err = p.recurse(&addr, &value, nil, nil, recurseLevel+1); err != nil {
			return err
		}
	}

	return nil
}
