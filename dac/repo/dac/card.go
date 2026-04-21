package dac

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
	"trpc.group/trpc-go/trpc-go/log"
)

// GetAllCards 获取指定模组下所有卡片（支持后置回调过滤）。
func GetAllCards(tx *gorm.DB, mozuID string, afterGet func(*gorm.DB, []db.Card) ([]db.Card, error)) ([]db.Card, error) {
	opts := make([]tgorm.Option, 0, 1)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)

	cards := make([]db.Card, 0)
	err := tgorm.WithOptions(tx, opts...).Find(&cards).Error

	if err != nil {
		return cards, err
	}

	if afterGet != nil {
		if cards, err = afterGet(tx, cards); err != nil {
			return cards, err
		}
	}

	return cards, nil
}

// GetAllCardsWithStaffAndAccessGroup 查询所有门禁卡信息，附带人员与权限组
// 返回值：
// - []db.Card: 所有卡信息
// - map[db.IDType]db.Staff: 员工ID到员工信息的映射
// - map[string][]db.IDType: 卡号到权限组ID列表的映射
// - map[db.IDType]string: 权限组ID到权限组名称的映射
func (d *impl) GetAllCardsWithStaffAndAccessGroup(ctx context.Context,
	mozuID string,
) ([]db.Card, map[db.IDType]db.Staff, map[string][]db.IDType,
	map[db.IDType]string, error,
) {
	var (
		cards        []db.Card
		staffMap     map[db.IDType]db.Staff
		cardGroupMap map[string][]db.IDType
		groupNameMap map[db.IDType]string
		err          error
	)

	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 查询所有卡信息
		cards, err = GetAllCards(tx, mozuID, nil)
		if err != nil {
			return fmt.Errorf("get all cards error: %w", err)
		}
		if len(cards) == 0 {
			staffMap = make(map[db.IDType]db.Staff)
			cardGroupMap = make(map[string][]db.IDType)
			groupNameMap = make(map[db.IDType]string)
			return nil
		}

		// 2. 查询所有卡对应的人员信息
		staffIDs := make([]db.IDType, 0, len(cards))
		cardNumbers := make([]string, 0, len(cards))
		for i := range cards {
			c := &cards[i]
			if c.StaffID != db.DefaultStaffID {
				staffIDs = append(staffIDs, c.StaffID)
			}
			cardNumbers = append(cardNumbers, c.CardNo)
		}
		staffMap, err = GetStaffsByID(tx, staffIDs)
		if err != nil {
			return fmt.Errorf("get staffs by id error: %w", err)
		}

		// 3. 查询卡与权限组的关联关系
		accessGroupRelations, err := GetCardAccessRelationByCards(tx, cardNumbers, mozuID)
		if err != nil {
			return fmt.Errorf("get card access group relation error: %w", err)
		}
		cardGroupMap = make(map[string][]db.IDType, len(accessGroupRelations))
		groupIDs := make(map[db.IDType]struct{}, len(accessGroupRelations))
		for i := range accessGroupRelations {
			groupID := accessGroupRelations[i].AccessGroupID
			cardNo := accessGroupRelations[i].CardNo

			groupIDs[groupID] = struct{}{}
			cardGroupMap[cardNo] = append(cardGroupMap[cardNo], groupID)
		}

		// 4. 查询权限组信息
		groupIDList := make([]db.IDType, 0, len(groupIDs))
		for id := range groupIDs {
			groupIDList = append(groupIDList, id)
		}

		groupBaseInfoMap, err := getAccessGroupBaseInfo(tx, groupIDList)
		if err != nil {
			return fmt.Errorf("get access group base info error: %w", err)
		}

		// 7. 构建权限组ID到名称的映射
		groupNameMap = make(map[db.IDType]string, len(groupBaseInfoMap))
		for i := range groupBaseInfoMap {
			g := &groupBaseInfoMap[i]
			groupNameMap[g.ID] = g.Name
		}

		return nil
	})

	if err != nil {
		return nil, nil, nil, nil, err
	}

	return cards, staffMap, cardGroupMap, groupNameMap, nil
}

// GetAllCards 获取所有卡片（impl 方法，支持后置回调过滤）。
func (d *impl) GetAllCards(ctx context.Context,
	afterGet func(*gorm.DB, []db.Card) ([]db.Card, error),
) ([]db.Card, error) {
	var (
		cards []db.Card
		e     error
	)
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cards, e = GetAllCards(tx, "", afterGet)
		return e
	})
	return cards, err
}

// GetCardsByStaffs 根据人员ID列表获取关联的卡片。
func GetCardsByStaffs(tx *gorm.DB, staffIDs []db.IDType) ([]db.Card, error) {
	cards := make([]db.Card, 0)
	if len(staffIDs) == 0 {
		return cards, nil
	}

	if err := tgorm.WithOptions(tx, withCardStaffIDs(staffIDs)).Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

// GetCardsByStaffs 根据人员ID列表获取关联的卡片（impl 方法）。
func (d *impl) GetCardsByStaffs(ctx context.Context, staffIDs []db.IDType) ([]db.Card, error) {
	return GetCardsByStaffs(d.db.WithContext(ctx), staffIDs)
}

// AddCard 添加卡片记录（校验人员ID后创建，支持后置回调）。
func (d *impl) AddCard(ctx context.Context, card db.Card, afterAdd func(*gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		staffID, err := getStaffID(tx, card.StaffID)
		if err != nil {
			return fmt.Errorf("get staff id error: %w", err)
		}
		card.StaffID = staffID

		if err = tx.Create(&card).Error; err != nil {
			return err
		}

		if afterAdd != nil {
			if err = afterAdd(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetCards 分页查询卡片列表（支持卡号、类型、标志、权限组等多条件过滤）。
func (d *impl) GetCards(ctx context.Context, mozuID string,
	cardNumbers []string, query string,
	cardType db.CardType, queryCardType bool,
	cardFlag db.CardFlagType, queryCardFlag bool,
	accessGroupID db.IDType, queryAccessGroup bool,
	offset int, limit int) (int64, []db.Card, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, errors.New(fmt.Sprintf("GetCards not support by offset:%v, limit:%v",
			offset, limit))
	}

	var (
		totalCount  int64
		cardRecords []db.Card
		e           error
		opts        = make([]tgorm.Option, 0, 10)
	)

	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	orOpts := make([]tgorm.Option, 0, 2)

	if len(cardNumbers) > 0 {
		opts = append(opts, withCardNumbersOption(cardNumbers))
	} else {
		orOpts = append(orOpts, withCardLike(query))
	}

	if queryCardFlag {
		opts = append(opts, withCardFlag(cardFlag))
	}

	if queryCardType {
		opts = append(opts, withCardType(cardType))
	}

	e = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var staffs []db.Staff
		if len(query) > 0 {
			if err := tgorm.WithOptions(tx, withNameLike(query)).Find(&staffs).Error; err != nil {
				return err
			}
			staffIDs := make([]db.IDType, 0, len(staffs))
			for i := range staffs {
				staffIDs = append(staffIDs, staffs[i].ID)
			}
			orOpts = append(orOpts, withCardStaffIDs(staffIDs))

			opts = append(opts, tgorm.WithOr(tx, orOpts...))
		}

		if queryAccessGroup {
			var relations []db.CardAccessRelation
			if err := tgorm.WithOptions(tx, withAccessGroupIDOption(accessGroupID)).Find(&relations).Error; err != nil {
				return err
			}
			queryCards := getCardsFromCardRelations(relations)
			opts = append(opts, withCardNumbersOption(queryCards))
		}

		return queryAndCountRecords(tgorm.WithOptions(tx.Model(&db.Card{}), opts...),
			offset, limit, &cardRecords, &totalCount)
	})

	return totalCount, cardRecords, e
}

// GetCardsByCardNos 根据卡号列表获取卡片信息。
func GetCardsByCardNos(tx *gorm.DB, cardNos []string, mozuID string) ([]db.Card, error) {
	cards := make([]db.Card, 0)
	if len(cardNos) == 0 {
		return cards, nil
	}

	err := tgorm.WithOptions(tx, withCardsMozuOption(cardNos, mozuID)...).Find(&cards).Error
	return cards, err
}

// GetCardsByCardNos 根据卡号列表获取卡片信息（impl 方法）。
func (d *impl) GetCardsByCardNos(ctx context.Context, cardNos []string, mozuID string) ([]db.Card, error) {
	return GetCardsByCardNos(d.db.WithContext(ctx), cardNos, mozuID)
}

// UnbindCards 批量解绑卡片与人员的关联（将 staff_id 设为默认值）。
func (d *impl) UnbindCards(ctx context.Context, cards []string, mozuID string) error {
	if len(cards) == 0 {
		return nil
	}

	return tgorm.WithOptions(
		d.db.WithContext(ctx).Model(&db.Card{}),
		withCardsMozuOption(cards, mozuID)...,
	).Updates(map[string]interface{}{
		db.ColumnStaffID: db.DefaultStaffID,
	}).Error
}

// UpdateCardsStaff 批量更新卡片关联的人员（校验人员ID后更新，支持后置回调）。
func (d *impl) UpdateCardsStaff(ctx context.Context, cards []string, staffID db.IDType, mozuID string,
	afterUpdate func(tx *gorm.DB) error) error {
	if len(cards) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newID, err := getStaffID(tx, staffID)
		if err != nil {
			return fmt.Errorf("invalid staff id: %v, error: %w", staffID, err)
		}
		if newID == db.DefaultStaffID {
			return fmt.Errorf("not found staff id: %v", staffID)
		}

		if err = tgorm.WithOptions(
			tx.Model(&db.Card{}),
			withCardsMozuOption(cards, mozuID)...,
		).Updates(map[string]interface{}{
			db.ColumnStaffID: staffID,
		}).Error; err != nil {
			return err
		}

		if afterUpdate != nil {
			if err = afterUpdate(tx); err != nil {
				return fmt.Errorf("after update cards %v staff %v error: %w", cards, staffID, err)
			}
		}
		return nil
	})
}

// UpdateCardsFlag 批量更新卡片标志（支持后置回调）。
func UpdateCardsFlag(tx *gorm.DB, cards []string, mozuID string,
	flag db.CardFlagType, afterUpdate func(*gorm.DB) error,
) error {
	if len(cards) == 0 {
		return nil
	}

	var err error

	if err = tgorm.WithOptions(tx.Model(&db.Card{}), withCardsMozuOption(cards, mozuID)...).Updates(map[string]interface{}{
		db.ColumnCardFlag: flag,
	}).Error; err != nil {
		return fmt.Errorf("update cards %v flag %v error: %w", cards, flag, err)
	}

	if afterUpdate != nil {
		if err = afterUpdate(tx); err != nil {
			return err
		}
	}
	return nil
}

// UpdateCardsType 批量更新卡片类型。
func UpdateCardsType(tx *gorm.DB, cards []string, mozuID string, cardType db.CardType) error {
	if len(cards) == 0 {
		return nil
	}
	var err error

	if err = tgorm.WithOptions(tx.Model(&db.Card{}), withCardsMozuOption(cards, mozuID)...).Updates(map[string]interface{}{
		db.ColumnCardType: cardType,
	}).Error; err != nil {
		return fmt.Errorf("update cards %v flag %v error: %w", cards, cardType, err)
	}

	return nil
}

// UpdateCardValidTime 批量更新卡片有效期（非未来时间或非永久卡时跳过下发）。
func UpdateCardValidTime(tx *gorm.DB, cards []string, mozuID string,
	validTime int64, afterUpdate func(*gorm.DB) error,
) error {
	var err error

	if err = tgorm.WithOptions(tx.Model(&db.Card{}), withCardsMozuOption(cards, mozuID)...).Updates(map[string]interface{}{
		db.ColumnCardValidTime: validTime,
	}).Error; err != nil {
		return fmt.Errorf("update cards %v flag %v error: %w", cards, validTime, err)
	}

	if afterUpdate != nil {
		// 判断是否需要续期
		if validTime <= time.Now().Unix() && validTime != 0 {
			log.Infof("更新有效期不是未来时间或不是永久卡，不需要续期，不涉及下发门控器，卡号：%v，续期时间：%v", cards, validTime)
			return nil
		}
		if err = afterUpdate(tx); err != nil {
			return err
		}
	}
	return nil
}

// UpdateCardsFlag 批量更新卡片标志（impl 方法，在事务中执行）。
func (d *impl) UpdateCardsFlag(ctx context.Context, cards []string,
	mozuID string, flag db.CardFlagType,
	afterUpdate func(*gorm.DB) error,
) error {
	if len(cards) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return UpdateCardsFlag(tx, cards, mozuID, flag, afterUpdate)
	})
}

// UpdateCardsType 批量更新卡片类型（impl 方法，在事务中执行）。
func (d *impl) UpdateCardsType(ctx context.Context, cards []string, mozuID string, cardType db.CardType) error {
	if len(cards) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return UpdateCardsType(tx, cards, mozuID, cardType)
	})
}

// UpdateCardValidTime 批量更新卡片有效期（impl 方法，在事务中执行）。
func (d *impl) UpdateCardValidTime(ctx context.Context, cards []string,
	mozuID string, validTime int64,
	afterUpdate func(*gorm.DB) error,
) error {
	if len(cards) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return UpdateCardValidTime(tx, cards, mozuID, validTime, afterUpdate)
	})
}

// DeleteCards 批量删除卡片及其权限组关联（支持前置回调）。
func DeleteCards(tx *gorm.DB, cards []string, mozuID string, beforeDelete func(*gorm.DB) error) error {
	if len(cards) == 0 {
		return nil
	}

	var err error
	if beforeDelete != nil {
		if err = beforeDelete(tx); err != nil {
			return err
		}
	}

	if err = tgorm.WithOptions(tx, withCardsMozuOption(cards, mozuID)...).Delete(&db.Card{}).Error; err != nil {
		return fmt.Errorf("delete mozu %v cards %v error: %w", mozuID, cards, err)
	}
	if err = deleteCardsAccessGroupRelation(tx, cards, mozuID); err != nil {
		return fmt.Errorf("delete mozu %v cards %v access group error: %w", mozuID, cards, err)
	}

	return nil
}

// DeleteCards 批量删除卡片（impl 方法，在事务中执行）。
func (d *impl) DeleteCards(ctx context.Context, cards []string,
	mozuID string, beforeDelete func(*gorm.DB) error,
) error {
	if len(cards) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return DeleteCards(tx, cards, mozuID, beforeDelete)
	})
}
