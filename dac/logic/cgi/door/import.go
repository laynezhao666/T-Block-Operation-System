// Package door 提供门的查询、更新、状态控制和导入导出功能。
package door

import (
	"context"
	"mime/multipart"

	"dac/entity/model/rt"
	"dac/logic/door"

	"dac/entity/utils/excel"
	"github.com/tealeg/xlsx/v3"
)

// parseSheet 解析Excel文件中的门编号数据
func parseSheet(file *xlsx.File) ([]rt.DoorWithCodeItem, error) {
	return excel.ParseFirstSheet[rt.DoorWithCodeItem](file, nil)
}

// Import 从Excel文件导入门编号数据
func Import(ctx context.Context, file *multipart.FileHeader) error {
	xf, err := excel.OpenFile(file)
	if err != nil {
		return err
	}

	items, err := parseSheet(xf)
	if err != nil {
		return err
	}

	return door.UpdateCode(ctx, items)
}
