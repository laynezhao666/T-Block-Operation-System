package parse

import (
	"fmt"
	"reflect"
)

func (p *parser) getDstStructFieldNameAddr(dstValue *reflect.Value, dstType reflect.Type) (map[string]reflect.Value,
	error) {
	n := dstType.NumField()
	dstName2Addr := make(map[string]reflect.Value, n)
	var ok bool
	for i := 0; i < n; i++ {
		dstFieldType := dstType.Field(i)
		if !dstFieldType.IsExported() {
			continue
		}
		if dstFieldType.Anonymous {
			newDstValue := dstValue.Field(i).Addr().Elem()
			newDstName2Addr, err := p.getDstStructFieldNameAddr(&newDstValue, newDstValue.Type())
			if err != nil {
				return nil, err
			}
			for k, v := range newDstName2Addr {
				if _, ok = dstName2Addr[k]; ok {
					return nil, fmt.Errorf("field \"%v\" is repeated", k)
				}
				dstName2Addr[k] = v
			}
		} else {
			dstFieldName, ok := p.config.FieldExtractor(&dstFieldType)
			if !ok {
				continue
			}

			dstFieldValue := dstValue.Field(i)
			dstName2Addr[dstFieldName] = dstFieldValue.Addr()
		}
	}
	return dstName2Addr, nil
}

func (p *parser) getSrcStructFieldNameValue(srcValue *reflect.Value, srcType reflect.Type,
	dstName2Addr map[string]reflect.Value) (map[string]reflect.Value, error) {
	n := srcType.NumField()
	srcName2Value := make(map[string]reflect.Value, n)
	var ok bool
	for i := 0; i < n; i++ {
		srcFieldType := srcType.Field(i)
		if !srcFieldType.IsExported() {
			continue
		}

		if srcFieldType.Anonymous {
			newSrcValue := srcValue.Field(i)
			newSrcName2Value, err := p.getSrcStructFieldNameValue(&newSrcValue, newSrcValue.Type(), dstName2Addr)
			if err != nil {
				return nil, err
			}
			for k, v := range newSrcName2Value {
				if _, ok = srcName2Value[k]; ok {
					return nil, fmt.Errorf("field \"%v\" is repeated", k)
				}
				if _, ok = dstName2Addr[k]; !ok {
					continue
				}
				srcName2Value[k] = v
			}
		} else {
			srcFieldName, ok := p.config.FieldExtractor(&srcFieldType)
			if !ok {
				continue
			}
			if _, ok = dstName2Addr[srcFieldName]; !ok {
				continue
			}
			srcName2Value[srcFieldName] = srcValue.Field(i)
		}
	}
	return srcName2Value, nil
}
