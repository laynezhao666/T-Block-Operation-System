package dac

import (
	"context"
	"fmt"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// delCardAccessRelByGroup 删除指定权限组的卡片关联关系。
func delCardAccessRelByGroup(tx *gorm.DB, id db.IDType) error {
	return withAccessGroupID(tx, id).Delete(&db.CardAccessRelation{}).Error
}

// GetCardAccessGroupRelationByCards 获取卡片与权限组的关联关系（impl 方法）。
func (d *impl) GetCardAccessGroupRelationByCards(ctx context.Context,
	cards []string, mozuID string,
) ([]db.CardAccessRelation, error) {
	return GetCardAccessRelationByCards(d.db.WithContext(ctx), cards, mozuID)
}

// getCardAccessGroupMap 从关联关系中构建卡号到权限组ID列表的映射。
func getCardAccessGroupMap(relations []db.CardAccessRelation) map[string][]db.IDType {
	m := make(map[string][]db.IDType, len(relations))
	for i := range relations {
		card := relations[i].CardNo
		m[card] = append(m[card], relations[i].AccessGroupID)
	}
	return m
}

// getCardsFromCardRelations 从关联关系中提取去重后的卡号列表。
func getCardsFromCardRelations(relations []db.CardAccessRelation) []string {
	cardMap := make(map[string]struct{})
	for i := range relations {
		cardMap[relations[i].CardNo] = struct{}{}
	}
	cards := make([]string, 0, len(cardMap))
	for card := range cardMap {
		cards = append(cards, card)
	}
	return cards
}

// getAccessGroupIDFromCardRelations 从关联关系中提取去重后的权限组ID列表。
func getAccessGroupIDFromCardRelations(cardRelations []db.CardAccessRelation) []db.IDType {
	idMap := make(map[db.IDType]struct{})
	for i := range cardRelations {
		idMap[cardRelations[i].AccessGroupID] = struct{}{}
	}

	ids := make([]int, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}
	return ids
}

// GetCardAccessGroupMapByGroupIDs 根据权限组ID列表获取卡号到权限组的映射。
func GetCardAccessGroupMapByGroupIDs(tx *gorm.DB, accessGroupIDs []db.IDType) (map[string][]db.IDType, []db.IDType, error) {
	if len(accessGroupIDs) == 0 {
		return nil, nil, nil
	}

	relations, err := GetCardAccessRelationByAccessGroups(tx, accessGroupIDs)
	if err != nil {
		return nil, nil, err
	}

	return getCardAccessGroupMap(relations), getAccessGroupIDFromCardRelations(relations), nil
}

// GetCardAccessGroupMap 根据卡号列表获取卡号到权限组的映射。
func GetCardAccessGroupMap(tx *gorm.DB, cards []string, mozuID string) (map[string][]db.IDType, []db.IDType, error) {
	if len(cards) == 0 {
		return nil, nil, nil
	}

	relations, err := GetCardAccessRelationByCards(tx, cards, mozuID)
	if err != nil {
		return nil, nil, err
	}

	return getCardAccessGroupMap(relations), getAccessGroupIDFromCardRelations(relations), nil
}

// GetAccessGroupIDByCards 根据卡号列表获取关联的权限组ID列表。
func GetAccessGroupIDByCards(tx *gorm.DB, cards []string, mozuID string) ([]db.IDType, error) {
	if len(cards) == 0 {
		return make([]db.IDType, 0), nil
	}

	cardRelations, err := GetCardAccessRelationByCards(tx, cards, mozuID)
	if err != nil {
		return nil, err
	}

	return getAccessGroupIDFromCardRelations(cardRelations), nil
}

// GetCardNumbersByAccessGroupIDs 根据权限组ID列表获取关联的卡号列表。
func GetCardNumbersByAccessGroupIDs(tx *gorm.DB, accessGroupIDs []db.IDType) ([]string, error) {
	if len(accessGroupIDs) == 0 {
		return make([]string, 0), nil
	}

	relations, err := GetCardAccessRelationByAccessGroups(tx, accessGroupIDs)
	if err != nil {
		return nil, err
	}

	return getCardsFromCardRelations(relations), nil
}

// GetCardAccessRelationByCards 根据卡号列表获取卡片与权限组的关联关系。
func GetCardAccessRelationByCards(tx *gorm.DB, cards []string, mozuID string) ([]db.CardAccessRelation, error) {
	if len(cards) == 0 {
		return nil, nil
	}

	var relations []db.CardAccessRelation
	err := tgorm.WithOptions(tx, withCardsMozuOption(cards, mozuID)...).Find(&relations).Error
	return relations, err
}

// GetCardAccessRelationByAccessGroups 根据权限组ID列表获取卡片与权限组的关联关系。
func GetCardAccessRelationByAccessGroups(tx *gorm.DB, accessGroupIDs []db.IDType) ([]db.CardAccessRelation, error) {
	if len(accessGroupIDs) == 0 {
		return make([]db.CardAccessRelation, 0), nil
	}

	var relations []db.CardAccessRelation
	err := withAccessGroupIDs(tx, accessGroupIDs).Find(&relations).Error
	return relations, err
}

// AddCardsAccessGroupRelation 添加数据库卡号与权限组的关系
func AddCardsAccessGroupRelation(tx *gorm.DB, cards []string, groups []db.IDType, mozuID string) error {
	l1 := len(cards)
	l2 := len(groups)
	if l1 == 0 || l2 == 0 {
		return nil
	}

	relations := make([]db.CardAccessRelation, 0, l1*l2)
	for i := range cards {
		for j := range groups {
			relations = append(relations, db.CardAccessRelation{
				CardNo:        cards[i],
				AccessGroupID: groups[j],
				MozuID:        mozuID,
			})
		}
	}

	return tx.Create(&relations).Error
}

// UpdateCardAccessGroupRelation 更新卡片与权限组的关联关系（impl 方法，支持前置和后置回调）。
func (d *impl) UpdateCardAccessGroupRelation(ctx context.Context, cards []string, groups []db.IDType, mozuID string,
	deleteIfEmpty bool, beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error {
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}
		if err = updateCardsAccessGroupRelation(tx, cards, groups, mozuID, deleteIfEmpty); err != nil {
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

// updateCardsAccessGroupRelation 更新卡片与权限组的关联关系（先校验权限组存在性，再先删后增）。
func updateCardsAccessGroupRelation(tx *gorm.DB, cards []string,
	groups []db.IDType, mozuID string, deleteIfEmpty bool,
) error {
	groupsLen := len(groups)
	// 避免误删除
	if groupsLen == 0 && !deleteIfEmpty {
		return nil
	}

	var (
		records []db.AccessGroup
	)

	err := queryRecordsByIDs(tx.Select("id"), groups, &records)
	if err != nil {
		return fmt.Errorf("query access group error: %w", err)
	}
	if len(records) != groupsLen {
		existIDs := make(map[db.IDType]struct{})
		for i := range records {
			existIDs[records[i].ID] = struct{}{}
		}
		missedIDs := make([]db.IDType, 0, groupsLen)
		for i := range groups {
			if _, ok := existIDs[groups[i]]; !ok {
				missedIDs = append(missedIDs, groups[i])
			}
		}
		if len(missedIDs) > 0 {
			return fmt.Errorf("access group %v not existed", missedIDs)
		}
	}

	cardsLen := len(cards)
	if cardsLen == 0 {
		return nil
	}

	if err = deleteCardsAccessGroupRelation(tx, cards, mozuID); err != nil {
		return err
	}

	return AddCardsAccessGroupRelation(tx, cards, groups, mozuID)
}

// updateAccessGroupCardsRelation 更新权限组与卡片的关联关系（先删后增）。
func updateAccessGroupCardsRelation(tx *gorm.DB, accessGroupID db.IDType, cards []string, mozuID string) error {
	// 删除原有记录
	if err := delCardAccessRelByGroup(tx, accessGroupID); err != nil {
		return err
	}

	if len(cards) == 0 { // 当卡信息为空时，只做删除
		return nil
	}

	// 新增
	cardAccessRelation := make([]db.CardAccessRelation, len(cards))
	for i := range cardAccessRelation {
		cardAccessRelation[i] = db.CardAccessRelation{
			CardNo:        cards[i],
			AccessGroupID: accessGroupID,
			MozuID:        mozuID,
		}
	}
	return tx.Create(&cardAccessRelation).Error
}
