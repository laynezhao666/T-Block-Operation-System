// Package staff 提供人员的增删改查、导入导出和密码管理功能。
package staff

import (
	"context"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils/excel"
	"dac/repo/dac"
	"fmt"
	"github.com/tealeg/xlsx/v3"
)

// ht 默认Excel行高
const (
	ht = 14.0
)

// titles 人员导出Excel表头
var (
	titles = []string{"姓名", "密码", "性别", "电话", "邮箱", "人员组", "证件类型", "证件号码", "备注"}
)

// writeExcel 将人员信息数据写入excel
func writeExcel(staffs []db.Staff) (*xlsx.File, error) {
	f := xlsx.NewFile()
	s, err := f.AddSheet(sheetName)
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range staffs {
		r, err := excel.AddStringRow(s, ht)
		if err != nil {
			return nil, err
		}

		staff := &staffs[i]

		if r.WriteStruct(&rt.StaffImportItem{
			Name:      staff.Name,
			Password:  staff.Password,
			Sex:       staff.Sex,
			Phone:     staff.Phone,
			Email:     staff.Email,
			Company:   staff.Company,
			PaperType: staff.PaperType,
			Paper:     staff.Paper,
			Comment:   staff.Comment,
		}, -1) < 0 {
			return nil, fmt.Errorf("write %+v error", *staff)
		}
	}

	return f, nil
}

// Export 导出模组下所有人员信息到Excel
func Export(ctx context.Context, mozuID string) (*xlsx.File, error) {
	staffs, err := dac.GetRW().GetAllStaffs(ctx, mozuID)
	if err != nil {
		return nil, fmt.Errorf("查询当前模组%v的人员信息失败：%v", mozuID, err)
	}

	return writeExcel(staffs)
}
