// Package staff 提供人员的增删改查、导入导出和密码管理功能。
package staff

import (
	"context"
	"dac/entity/consts"
	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/logic/cache"
	"dac/logic/card"
	"dac/repo/dac"
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

// passwordRegex 人员密码格式校验正则（6位纯数字）
var (
	passwordRegex = regexp.MustCompile(`^[[:digit:]]{6}$`)
)

// GetAllStaffCompany 获取模组下所有人员的单位列表
func GetAllStaffCompany(ctx context.Context, mozuID string) ([]string, error) {
	return dac.GetRW().GetAllStaffCompany(ctx, mozuID)
}

// GetStaffs 分页查询人员列表，支持按关键字和单位过滤
func GetStaffs(ctx context.Context, mozuID string, offset int, limit int,
	query string, company string) (cgi.Staffs, error) {
	total, dbStaffs, err := dac.GetRW().GetStaffs(ctx, mozuID, offset, limit, query, company)
	if err != nil {
		return cgi.Staffs{}, fmt.Errorf("GetStaffs error: %w", err)
	}

	// 获取人员关联的卡
	staffIDs := make([]db.IDType, len(dbStaffs))
	for i := range dbStaffs {
		staffIDs[i] = dbStaffs[i].ID
	}
	cards, err := dac.GetRW().GetCardsByStaffs(ctx, staffIDs)
	if err != nil {
		return cgi.Staffs{}, fmt.Errorf("GetCardsByStaffs error: %w", err)
	}

	// 构建人员和对应卡集合的Map关系
	staffRelateCards := make(map[db.IDType][]string) // 人员关联的卡
	for i := range cards {
		staffRelateCards[cards[i].StaffID] = append(staffRelateCards[cards[i].StaffID], cards[i].CardNo)
	}

	// 构建cgi对象
	var staffs cgi.Staffs
	staffs.Total = total
	staffs.List = make([]cgi.StaffAndCard, len(dbStaffs))
	for i := range dbStaffs {
		staffs.List[i] = cgi.StaffAndCard{
			Staff:  dbStaffs[i],
			CardNo: staffRelateCards[dbStaffs[i].ID],
		}

		utils.ProcessPersonalInformation(&staffs.List[i].Staff)
	}

	return staffs, err
}

// fillDefaultStaffPassword 为空密码的人员填充默认密码
func fillDefaultStaffPassword(staff *db.Staff) {
	if len(staff.Password) > 0 {
		return
	}
	staff.Password = consts.DefaultPassword
}

// verifyPassword 校验密码格式是否为6位数字
func verifyPassword(p string) bool {
	if len(p) == 0 {
		return true
	}
	return passwordRegex.MatchString(p)
}

// verifyStaff 校验人员数据的有效性
func verifyStaff(staff *db.Staff) error {
	if !verifyPassword(staff.Password) {
		return fmt.Errorf("password %v must be 6 digits", staff.Password)
	}
	return nil
}

// AddStaff 添加单个人员
func AddStaff(ctx context.Context, mozuID string, staff db.Staff) (db.IDType, error) {
	// 验证用户密码，密码必须为6位数字，可以为空，但密码为空时，会使用默认密码
	if err := verifyStaff(&staff); err != nil {
		return 0, err
	}
	fillDefaultStaffPassword(&staff)

	staff.MozuID = mozuID
	return dac.GetRW().AddStaff(ctx, staff)
}

// AddStaffs 批量添加人员
func AddStaffs(ctx context.Context, staffs []db.Staff) error {
	// 验证用户密码，密码必须为6位数字，可以为空，但密码为空时，会使用默认密码
	for _, staff := range staffs {
		if err := verifyStaff(&staff); err != nil {
			return err
		}
		fillDefaultStaffPassword(&staff)
	}

	return dac.GetRW().AddStaffs(ctx, staffs)
}

// UpdateStaff 更新人员信息，密码变更时同步到门禁控制器
func UpdateStaff(ctx context.Context, id db.IDType, staff db.Staff, mozuID string) error {
	err := verifyStaff(&staff)
	if err != nil {
		return err
	}

	var oldStaffData db.Staff
	return dac.GetRW().UpdateStaff(ctx, id, staff, func(tx *gorm.DB) error {
		oldStaffData, err = dac.GetStaffByID(tx, id)
		return err
	}, func(tx *gorm.DB) error {
		currentStaff, err := dac.GetStaffByID(tx, id)
		if err != nil {
			return err
		}
		cache.Get().UpdateStaff(currentStaff, mozuID)

		needUpdateInController := false
		// 仅当人员密码更新时，才会下发更新信息到对应门禁控制器
		if len(staff.Password) > 0 && staff.Password != oldStaffData.Password {
			needUpdateInController = true
		}
		if !needUpdateInController {
			return nil
		}

		if len(staff.Name) == 0 {
			staff.Name = oldStaffData.Name
		}
		if len(staff.Password) == 0 {
			staff.Password = oldStaffData.Password
		}

		cards, err := dac.GetCardsByStaffs(tx, []db.IDType{id})
		if err != nil {
			return err
		}

		cardNumbers := make([]string, 0, len(cards))
		for i := range cards {
			cardNumbers = append(cardNumbers, cards[i].CardNo)
		}

		return card.UpdateStaffInController(tx, cardNumbers, mozuID, staff.Name, staff.Password, int(id))
	})
}

// DeleteStaff 删除人员，支持三种模式：仅删人员/禁用卡/删除卡
func DeleteStaff(ctx context.Context, id db.IDType, deleteMode cgi.DeleteStaffModeType) error {
	switch deleteMode {
	case cgi.DeleteStaffOnly, cgi.DeleteStaffAndDisableCard, cgi.DeleteStaffAndDeleteCard:
	default:
		return fmt.Errorf("not support delete staff mode: %v", deleteMode)
	}

	var (
		cards []db.Card
		err   error
	)
	return dac.GetRW().DeleteStaff(ctx, id, func(tx *gorm.DB) error {
		cards, err = dac.GetCardsByStaffs(tx, []db.IDType{id})
		if err != nil {
			return err
		}
		return err
	}, func(tx *gorm.DB) error {
		mozuID := ""
		cardNumbers := make([]string, 0, len(cards))
		for i := range cards {
			cardNumbers = append(cardNumbers, cards[i].CardNo)
			mozuID = cards[i].MozuID
		}

		cache.Get().DeleteStaff(mozuID, id)

		switch deleteMode {
		case cgi.DeleteStaffOnly:
			// 删除人员后，不影响其拥有的门禁卡权限
			// 仅更新对应的用户名和密码
			return card.UpdateStaffInController(tx, cardNumbers, mozuID, consts.DefaultUserName, consts.DefaultPassword, int(id))
		case cgi.DeleteStaffAndDisableCard:
			// 删除人员后，禁用其拥有的门禁卡
			return card.UpdateFlagInTransaction(tx, cardNumbers, mozuID, db.CardFlagDisable)
		case cgi.DeleteStaffAndDeleteCard:
			return card.DeleteInTransaction(tx, cardNumbers, mozuID)
		}
		return nil
	})
}
