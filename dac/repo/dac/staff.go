package dac

import (
	"context"
	"fmt"
	"sort"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// GetStaffs 分页查询人员列表（支持模糊搜索姓名、手机、邮箱、备注，以及公司过滤）。
func (d *impl) GetStaffs(ctx context.Context, mozuID string, offset int, limit int,
	query string, company string) (int64, []db.Staff, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("GetStaffs not support by offset: %v, limit: %v",
			offset, limit)
	}

	var (
		totalCount int64
		staffs     []db.Staff
		opts       = make([]tgorm.Option, 0, 2)
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = append(opts, withNotStaffType(db.StaffTypeDeleted))

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(query) > 0 {
			opts = append(opts, tgorm.WithOr(tx, withNameLike(query), withPhoneLike(query),
				withEmailLike(query), withCommentLike(query)))
		}
		opts = addCompanyOptionIfNotEmpty(opts, company)
		return queryAndCountRecords(tgorm.WithOptions(tx.Model(&db.Staff{}), opts...), offset, limit, &staffs, &totalCount)
	})
	return totalCount, staffs, err
}

// GetAllStaffs 获取指定模组下所有人员（排除已删除）。
func (d *impl) GetAllStaffs(ctx context.Context, mozuID string) ([]db.Staff, error) {
	var (
		staffs []db.Staff
		opts   = make([]tgorm.Option, 0, 2)
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = append(opts, withNotStaffType(db.StaffTypeDeleted))

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tgorm.WithOptions(tx.Model(&db.Staff{}), opts...).Find(&staffs).Error
	})
	return staffs, err
}

// GetStaffsByID 根据ID列表获取人员映射（impl 方法）。
func (d *impl) GetStaffsByID(ctx context.Context, ids []db.IDType) (map[db.IDType]db.Staff, error) {
	return GetStaffsByID(d.db.WithContext(ctx), ids)
}

// AddStaff 添加单个人员记录，返回新建ID。
func (d *impl) AddStaff(ctx context.Context, staff db.Staff) (db.IDType, error) {
	tx := d.db.WithContext(ctx).Create(&staff)
	return staff.ID, tx.Error
}

// AddStaffs 批量添加人员记录。
func (d *impl) AddStaffs(ctx context.Context, staffs []db.Staff) error {
	tx := d.db.WithContext(ctx).Create(&staffs)
	return tx.Error
}

// UpdateStaff 更新人员信息（支持前置和后置回调）。
func (d *impl) UpdateStaff(ctx context.Context, id db.IDType, staff db.Staff,
	beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error,
) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error

		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		if err = withID(tx, id).Updates(&staff).Error; err != nil {
			return err
		}

		if afterUpdate != nil {
			if err = afterUpdate(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteStaff 软删除人员（标记为已删除类型，解绑关联卡片的 staff_id）。
func (d *impl) DeleteStaff(ctx context.Context, id db.IDType,
	beforeDelete func(*gorm.DB) error, afterDelete func(*gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		if beforeDelete != nil {
			if err = beforeDelete(tx); err != nil {
				return err
			}
		}
		// 软删除
		if err = withID(tx.Model(&db.Staff{}), id).Update("type", db.StaffTypeDeleted).Error; err != nil {
			return err
		}
		if err = withEqual(tx.Model(&db.Card{}), "staff_id", id).Update("staff_id", db.DefaultStaffID).Error; err != nil {
			return err
		}
		if afterDelete != nil {
			if err = afterDelete(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetAllStaffCompany 获取指定模组下所有人员的公司列表（去重排序）。
func (d *impl) GetAllStaffCompany(ctx context.Context, mozuID string) ([]string, error) {
	var (
		companies []string
		staffs    []db.Staff
		opts      = make([]tgorm.Option, 0, 2)
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = append(opts, withNotStaffType(db.StaffTypeDeleted))

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tgorm.WithOptions(tx, opts...).Find(&staffs).Error; e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	companyMap := make(map[string]struct{}, len(staffs))
	for i := range staffs {
		companyMap[staffs[i].Company] = struct{}{}
	}

	companies = make([]string, 0, len(companyMap))
	for c := range companyMap {
		companies = append(companies, c)
	}
	sort.Strings(companies)

	return companies, nil
}

// getStaffID 判断是否存在目标id的staff记录并返回id
func getStaffID(tx *gorm.DB, id db.IDType) (db.IDType, error) {
	var (
		count int64 = 0
		err   error
	)

	if err = countRecord(withID(tx.Model(&db.Staff{}), id), &count); err != nil {
		return db.DefaultStaffID, err
	}

	if count == 0 {
		id = db.DefaultStaffID
	}
	return id, nil
}

// GetStaffByID 根据ID获取单个人员记录。
func GetStaffByID(tx *gorm.DB, id db.IDType) (db.Staff, error) {
	var s db.Staff
	err := queryRecordByID(tx, id, &s)
	return s, err
}

// GetStaffsByID 根据ID列表获取人员映射。
func GetStaffsByID(tx *gorm.DB, ids []db.IDType) (map[db.IDType]db.Staff, error) {
	if len(ids) == 0 {
		return make(map[db.IDType]db.Staff), nil
	}

	var staffs []db.Staff
	err := queryRecordsByIDs(tx, ids, &staffs)
	if err != nil {
		return nil, err
	}

	results := make(map[db.IDType]db.Staff, len(staffs))
	for i := range staffs {
		results[staffs[i].ID] = staffs[i]
	}
	return results, nil
}

// getStaffsByCardNos 根据卡号列表获取关联的人员基本信息。
func getStaffsByCardNos(tx *gorm.DB, cardNos []string, mozuID string) ([]db.CardAndStaffBase, error) {
	if len(cardNos) == 0 {
		return nil, nil
	}

	// 查询卡号对应的人员ID
	cards, err := GetCardsByCardNos(tx, cardNos, mozuID)
	if err != nil {
		return nil, err
	}

	staffIDs := getCardsStaffID(cards)
	// 查询人员
	var staffs []db.StaffBase
	if err = queryRecordsByIDs(tx.Model(&db.Staff{}), staffIDs, &staffs); err != nil {
		return nil, err
	}
	staffMap := convertStaffsToMap(staffs)

	// 构建卡和人员信息
	cardAndStaff := make([]db.CardAndStaffBase, len(cards))
	for i := range cards {
		cardAndStaff[i].CardNo = cards[i].CardNo
		cardAndStaff[i].Staff = staffMap[cards[i].StaffID]
	}

	return cardAndStaff, nil
}

// convertStaffsToMap 将人员基本信息列表转换为以ID为key的map。
func convertStaffsToMap(staffs []db.StaffBase) map[db.IDType]db.StaffBase {
	staffMap := make(map[db.IDType]db.StaffBase)
	for i := range staffs {
		staffMap[staffs[i].ID] = staffs[i]
	}
	return staffMap
}
