package excel

import (
	"fmt"
	"mime/multipart"

	"github.com/tealeg/xlsx/v3"
)

type TraverseFunction func(r *xlsx.Row, index int) error

func ForEachRow(s *xlsx.Sheet, startRow int, pointer interface{}, f TraverseFunction) error {
	return s.ForEachRow(func(r *xlsx.Row) error {
		index := r.GetCoordinate()
		if index < startRow {
			return nil
		}

		if err := r.ReadStruct(pointer); err != nil {
			return err
		}
		return f(r, index)
	})
}

func GetSheet(f *xlsx.File, sheetName string) (*xlsx.Sheet, error) {
	s, ok := f.Sheet[sheetName]
	if !ok || s == nil {
		return nil, fmt.Errorf("\"%v\" sheet 不存在", sheetName)
	}
	return s, nil
}

// ParseFirstSheet 通用解析Excel第一个Sheet的泛型函数
// skipFunc 用于判断是否跳过当前行，传 nil 表示不跳过任何行
func ParseFirstSheet[T any](file *xlsx.File, skipFunc func(item T) bool) ([]T, error) {
	if len(file.Sheets) == 0 {
		return nil, fmt.Errorf("empty sheets")
	}

	s := file.Sheets[0]
	if s == nil {
		return nil, fmt.Errorf("nil sheet")
	}

	var (
		item    T
		err     error
		results = make([]T, 0, s.MaxRow)
	)

	if err = ForEachRow(s, 1, &item, func(r *xlsx.Row, index int) error {
		if skipFunc != nil && skipFunc(item) {
			return nil
		}
		results = append(results, item)
		return nil
	}); err != nil {
		return nil, err
	}

	return results, nil
}

// OpenFile 打开上传的multipart文件并解析为xlsx.File
func OpenFile(file *multipart.FileHeader) (*xlsx.File, error) {
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload file %v error: %w", file.Filename, err)
	}
	defer func() {
		_ = f.Close()
	}()

	xf, err := xlsx.OpenReaderAt(f, file.Size)
	if err != nil {
		return nil, fmt.Errorf("open excel file %v error: %w", file.Filename, err)
	}

	return xf, nil
}
