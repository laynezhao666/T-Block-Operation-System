// Package staff 提供人员的增删改查、导入导出和密码管理功能。
package staff

import (
	"context"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils/excel"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"mime/multipart"
	"strings"
)

// sheetName 人员导入Excel的Sheet名称
const (
	sheetName = "人员信息"
)

// convert 转换为staff结构体
func convert(item rt.StaffImportItem, mozu string) db.Staff {
	staff := db.Staff{
		StaffBase: db.StaffBase{
			Name: strings.TrimSpace(item.Name),
		},
		Password:  item.Password,
		Sex:       item.Sex,
		Phone:     item.Phone,
		Email:     item.Email,
		Company:   item.Company,
		PaperType: item.PaperType,
		Paper:     item.Paper,
		Comment:   item.Comment,
		MozuID:    mozu,
	}

	return staff
}

// parseSheet 解析excel中的人员数据
func parseSheet(file *xlsx.File) ([]rt.StaffImportItem, error) {
	return excel.ParseFirstSheet(file, func(item rt.StaffImportItem) bool {
		return isEmptyStaffItem(item)
	})
}

// isEmptyStaffItem 判断员工记录是否为空
func isEmptyStaffItem(item rt.StaffImportItem) bool {
	// 检查关键字段是否都为空
	return strings.TrimSpace(item.Name) == ""
}

// Import 从Excel文件批量导入人员数据
func Import(ctx context.Context, mozuID string, file *multipart.FileHeader) error {
	xf, err := excel.OpenFile(file)
	if err != nil {
		return err
	}

	records, err := parseSheet(xf)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("Excel文件中没有有效的人员数据")
	}

	// 转换为数据库结构体
	staffs := make([]db.Staff, len(records))
	for i, item := range records {
		staffs[i] = convert(item, mozuID)
	}
	return AddStaffs(ctx, staffs)
}
