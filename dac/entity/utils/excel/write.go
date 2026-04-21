// Package excel 提供Excel文件的读写工具函数。
package excel

import (
	"errors"
	"fmt"

	"github.com/tealeg/xlsx/v3"
)

// defaultHeight 默认行高
// centerAlignment 居中对齐方式
const (
	defaultHeight = 14.0

	centerAlignment = "center"
)

// AddCells 向指定行追加多个单元格，值自动转为字符串
func AddCells(r *xlsx.Row, cells ...interface{}) (*xlsx.Row, error) {
	if r == nil {
		return nil, errors.New("nil row")
	}

	for i := range cells {
		r.AddCell().SetString(fmt.Sprintf("%+v", cells[i]))
	}
	return r, nil
}

// AddStringCells 向指定行追加多个字符串单元格
func AddStringCells(r *xlsx.Row, cells ...string) (*xlsx.Row, error) {
	if r == nil {
		return nil, errors.New("nil row")
	}

	for _, c := range cells {
		r.AddCell().SetString(c)
	}
	return r, nil
}

// AddRow 向Sheet追加一行并填充数据，支持自定义行高
func AddRow(s *xlsx.Sheet, height float64, cells ...interface{}) (*xlsx.Row, error) {
	if s == nil {
		return nil, errors.New("nil sheet")
	}

	r := s.AddRow()
	if height <= 0.0 {
		height = defaultHeight
	}
	r.SetHeight(height)

	return AddCells(r, cells...)
}

// AddStringRow 向Sheet追加一行并填充字符串数据
func AddStringRow(s *xlsx.Sheet, height float64, cells ...string) (*xlsx.Row, error) {
	if s == nil {
		return nil, errors.New("nil sheet")
	}

	r := s.AddRow()
	if height <= 0.0 {
		height = defaultHeight
	}
	r.SetHeight(height)

	return AddStringCells(r, cells...)
}

// AddMergedCell 创建合并单元格并设置居中对齐
func AddMergedCell(c *xlsx.Cell, vertical, horizontal int, v string) error {
	if c == nil {
		return errors.New("nil cell")
	}

	c.Merge(horizontal, vertical)
	c.SetString(v)

	s := c.GetStyle()
	s.Alignment.Horizontal = centerAlignment
	s.Alignment.Vertical = centerAlignment
	c.SetStyle(s)

	return nil
}
