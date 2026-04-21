// Package dac 提供门禁系统的数据访问层，封装数据库CRUD操作。
package dac

import (
	"context"
	"fmt"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/rt"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// GetCardAndStaffsByAccessGroupIDs 根据权限组ID列表获取关联的卡片和人员信息。
func (d *impl) GetCardAndStaffsByAccessGroupIDs(ctx context.Context, ids []db.IDType, mozuID string) (
	map[db.IDType][]db.CardAndStaffBase, error) {
	cardAndStaffsByAccessGroup := make(map[db.IDType][]db.CardAndStaffBase)
	if len(ids) == 0 {
		return cardAndStaffsByAccessGroup, nil
	}

	// 查询权限组与卡ID的关联表
	accessAndCards, err := GetCardAccessRelationByAccessGroups(d.db.WithContext(ctx), ids)
	if err != nil {
		return cardAndStaffsByAccessGroup, nil
	}
	cardNos := getCardNosFromAccessGroupRelation(accessAndCards)

	// 根据卡号查询人员信息
	cardAndStaffs, err := getStaffsByCardNos(d.db.WithContext(ctx), cardNos, mozuID)
	if err != nil {
		return cardAndStaffsByAccessGroup, err
	}
	cardAndStaffMap := make(map[string]db.CardAndStaffBase)
	for i := range cardAndStaffs {
		c := &cardAndStaffs[i]
		cardAndStaffMap[c.CardNo] = *c
	}

	// 将卡和人员信息赋值到权限组map
	for i := range accessAndCards {
		r := &accessAndCards[i]
		c, ok := cardAndStaffMap[r.CardNo]
		if ok {
			cardAndStaffsByAccessGroup[r.AccessGroupID] = append(cardAndStaffsByAccessGroup[r.AccessGroupID], c)
		}
	}
	return cardAndStaffsByAccessGroup, nil
}

// createControllers 批量创建控制器记录
func createControllers(tx *gorm.DB, controllers []db.DoorController) error {
	if len(controllers) == 0 {
		return nil
	}

	return tx.Create(&controllers).Error
}

// getDoorControllers 根据ID列表获取控制器记录
func getDoorControllers(tx *gorm.DB, ids []db.IDType) ([]db.DoorController, error) {
	if len(ids) == 0 {
		return make([]db.DoorController, 0), nil
	}

	var controllers []db.DoorController
	err := queryRecordsByIDs(tx, ids, &controllers)
	return controllers, err
}

// deleteCardsAccessGroupRelation 删除指定卡号的权限组关联
func deleteCardsAccessGroupRelation(tx *gorm.DB, cards []string, mozuID string) error {
	if len(cards) == 0 {
		return nil
	}

	return tgorm.WithOptions(tx, withCardsMozuOption(cards, mozuID)...).Delete(&db.CardAccessRelation{}).Error
}

// getAccessGroupBaseInfo 根据ID列表获取权限组基础信息
func getAccessGroupBaseInfo(tx *gorm.DB, ids []db.IDType) ([]db.AccessGroupBaseInfo, error) {
	if len(ids) == 0 {
		return make([]db.AccessGroupBaseInfo, 0), nil
	}

	var groups []db.AccessGroupBaseInfo
	err := queryRecordsByIDs(tx.Model(&db.AccessGroup{}), ids, &groups)
	return groups, err
}

// GetControllerIDsByCards 根据卡号列表获取关联的控制器ID映射
func GetControllerIDsByCards(tx *gorm.DB, cards []string, mozuID string) (map[string]map[db.IDType]struct{}, error) {
	cardAccessGroupMap, accessGroupIDs, err := GetCardAccessGroupMap(tx, cards, mozuID)
	if err != nil {
		return nil, err
	}
	if len(accessGroupIDs) == 0 {
		return nil, nil
	}

	accessGroupControllerDoors, err := GetControllerDoorsByAccessGroups(tx, accessGroupIDs)
	if err != nil {
		return nil, err
	}
	if len(accessGroupControllerDoors) == 0 {
		return nil, nil
	}

	results := make(map[string]map[db.IDType]struct{}, len(cards))
	for _, card := range cards {
		results[card] = make(map[db.IDType]struct{})
		for _, accessGroupID := range cardAccessGroupMap[card] {
			for controllerID := range accessGroupControllerDoors[accessGroupID] {
				results[card][controllerID] = struct{}{}
			}
		}
	}
	return results, nil
}

// GetControllerIDsByAccessGroupIDs 根据权限组ID列表获取关联的控制器ID
func GetControllerIDsByAccessGroupIDs(tx *gorm.DB, accessGroupIDs []int) ([]db.IDType, error) {
	controllerIDs := make([]db.IDType, 0)
	if len(accessGroupIDs) == 0 {
		return controllerIDs, nil
	}

	doorsMap, err := GetAccessGroupsDoors(tx, accessGroupIDs)
	if err != nil {
		return controllerIDs, fmt.Errorf("get access group %v doors error: %w", accessGroupIDs, err)
	}

	controllerIDMap := make(map[db.IDType]struct{})
	for _, doors := range doorsMap {
		for i := range doors {
			controllerIDMap[doors[i].ControllerID] = struct{}{}
		}
	}

	controllerIDs = make([]db.IDType, 0, len(controllerIDs))
	for id := range controllerIDMap {
		controllerIDs = append(controllerIDs, id)
	}
	return controllerIDs, nil
}

// GetControllerDoorsByAccessGroups 根据权限组ID获取控制器-门映射关系
func GetControllerDoorsByAccessGroups(tx *gorm.DB, accessGroupIDs []int) (map[db.IDType]map[db.IDType][]int, error) {
	doorsMap, err := GetAccessGroupsDoors(tx, accessGroupIDs)
	if err != nil {
		return nil, fmt.Errorf("get access group %v doors error: %w", accessGroupIDs, err)
	}

	results := make(map[db.IDType]map[db.IDType][]int, len(accessGroupIDs))
	for groupID, doors := range doorsMap {
		results[groupID] = make(map[db.IDType][]int)
		for i := range doors {
			d := &doors[i]
			c := d.ControllerID

			results[groupID][c] = append(results[groupID][c], d.Number)
		}
	}

	return results, nil
}

// getCardControllerTimeGroupAndDoors 获取卡-控制器-时间组-门的完整映射关系
func getCardControllerTimeGroupAndDoors(tx *gorm.DB,
	cardAccessGroups map[string][]db.IDType, accessGroupIDs []db.IDType) (map[string]map[db.IDType]int,
	map[string]map[db.IDType]map[int]struct{}, map[string]map[db.IDType]map[db.IDType]struct{}, error) {
	accessGroupDoors, err := GetAccessGroupsDoors(tx, accessGroupIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	accessGroups, err := GetAccessGroupMapByID(tx, accessGroupIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get access gropus %v error: %w", accessGroupIDs, err)
	}

	cardControllerAccessGroups := make(map[string]map[db.IDType]map[int]struct{})
	cardControllerTimeGroups := make(map[string]map[db.IDType]int)
	cardControllerDoors := make(map[string]map[db.IDType]map[int]struct{})
	for card, accessGroupIDs := range cardAccessGroups {
		cardControllerTimeGroups[card] = make(map[db.IDType]int)
		cardControllerAccessGroups[card] = make(map[db.IDType]map[db.IDType]struct{})
		cardControllerDoors[card] = make(map[db.IDType]map[int]struct{})

		for _, accessGroupID := range accessGroupIDs {
			accessGroup, ok := accessGroups[accessGroupID]
			if !ok {
				return nil, nil, nil, fmt.Errorf("未获取到权限组 %v", accessGroupID)
			}
			doors, ok := accessGroupDoors[accessGroupID]
			if !ok {
				continue
			}

			for i := range doors {
				d := &doors[i]
				controllerID := d.ControllerID

				if _, ok := cardControllerAccessGroups[card][controllerID]; !ok {
					cardControllerAccessGroups[card][controllerID] = make(map[db.IDType]struct{})
				}
				cardControllerAccessGroups[card][controllerID][accessGroupID] = struct{}{}

				// 每个门禁控制器下的时间组必须一致
				if old, ok := cardControllerTimeGroups[card][controllerID]; ok && old != accessGroup.TimeGroupNo {
					return nil, nil, nil, fmt.Errorf("card \"%v\", controller %v has multiply time groups: %v, %v",
						card, controllerID, old, accessGroup.TimeGroupNo)
				}
				cardControllerTimeGroups[card][controllerID] = accessGroup.TimeGroupNo
				if _, ok := cardControllerDoors[card][controllerID]; !ok {
					cardControllerDoors[card][controllerID] = make(map[int]struct{})
				}
				cardControllerDoors[card][controllerID][d.Number] = struct{}{}
			}
		}
	}
	return cardControllerTimeGroups, cardControllerDoors, cardControllerAccessGroups, nil
}

// GetCardCtrlTimeGroupDoorsByGroups 根据权限组ID获取卡-控制器-时间组-门映射
func GetCardCtrlTimeGroupDoorsByGroups(tx *gorm.DB, accessGroupIDs []db.IDType) (
	map[string]map[db.IDType]int, map[string]map[db.IDType]map[int]struct{},
	map[string]map[db.IDType]map[db.IDType]struct{}, error) {
	cardAccessGroups, accessGroupIDs, err := GetCardAccessGroupMapByGroupIDs(tx, accessGroupIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return getCardControllerTimeGroupAndDoors(tx, cardAccessGroups, accessGroupIDs)
}

// GetCardCtrlTimeGroupDoorsByCards 根据卡号获取卡-控制器-时间组-门映射
func GetCardCtrlTimeGroupDoorsByCards(tx *gorm.DB, cards []string, mozuID string) (
	map[string]map[db.IDType]int, map[string]map[db.IDType]map[int]struct{},
	map[string]map[db.IDType]map[db.IDType]struct{}, error) {
	cardAccessGroups, accessGroupIDs, err := GetCardAccessGroupMap(tx, cards, mozuID)
	if err != nil {
		return nil, nil, nil, err
	}

	return getCardControllerTimeGroupAndDoors(tx, cardAccessGroups, accessGroupIDs)
}

// GetControllerTimeGroupAndDoors 根据权限组ID获取控制器-时间组-门映射
func GetControllerTimeGroupAndDoors(tx *gorm.DB, accessGroupIDs []int) (
	map[db.IDType]int, map[db.IDType]map[int]struct{}, error,
) {
	accessGroupDoorsMap, err := GetAccessGroupsDoors(tx, accessGroupIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("get access group %v doors error: %w", accessGroupIDs, err)
	}

	accessGroups, err := GetAccessGroupMapByID(tx, accessGroupIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("get access gropus %v error: %w", accessGroupIDs, err)
	}

	controllerTimeGroups := make(map[db.IDType]int)
	controllerDoorNumbers := make(map[db.IDType]map[int]struct{})
	for accessGroupID, doors := range accessGroupDoorsMap {
		accessGroup, ok := accessGroups[accessGroupID]
		if !ok {
			return nil, nil, fmt.Errorf("未获取到权限组 %v", accessGroupID)
		}
		for i := range doors {
			d := &doors[i]
			controllerID := d.ControllerID
			// 每个门禁控制器下的时间组必须一致
			if old, ok := controllerTimeGroups[controllerID]; ok && old != accessGroup.TimeGroupNo {
				return nil, nil, fmt.Errorf("controller %v has multiply time groups: %v, %v",
					controllerID, old, accessGroup.TimeGroupNo)
			}

			controllerTimeGroups[controllerID] = accessGroup.TimeGroupNo
			if _, ok := controllerDoorNumbers[controllerID]; !ok {
				controllerDoorNumbers[controllerID] = make(map[int]struct{})
			}
			controllerDoorNumbers[controllerID][d.Number] = struct{}{}
		}
	}

	return controllerTimeGroups, controllerDoorNumbers, nil
}

// getCardsStaffID 从卡列表中提取人员ID（去重）
func getCardsStaffID(cards []db.Card) []db.IDType {
	staffIDMap := make(map[db.IDType]struct{}, len(cards))
	for i := range cards {
		staffIDMap[cards[i].StaffID] = struct{}{}
	}

	staffIDs := make([]db.IDType, 0, len(staffIDMap))
	for id := range staffIDMap {
		staffIDs = append(staffIDs, id)
	}
	return staffIDs
}

// GetCardStaffMapByCards 根据卡号获取卡列表和卡号-人员映射
func GetCardStaffMapByCards(tx *gorm.DB, cardNumbers []string, mozuID string) ([]db.Card, map[string]db.Staff, error) {
	var (
		cards       []db.Card
		cardToStaff map[string]db.Staff
		err         error
	)

	if cards, err = GetCardsByCardNos(tx, cardNumbers, mozuID); err != nil {
		return nil, nil, err
	}

	staffIDs := getCardsStaffID(cards)
	staffs, err := GetStaffsByID(tx, staffIDs)
	if err != nil {
		return nil, nil, err
	}

	cardToStaff = make(map[string]db.Staff, len(cards))
	for i := range cards {
		t, ok := staffs[cards[i].StaffID]
		if !ok {
			continue
		}
		cardToStaff[cards[i].CardNo] = t
	}
	return cards, cardToStaff, nil
}

// GetAllCardStaffMap 获取所有模组的卡列表和卡号-人员映射
func (d *impl) GetAllCardStaffMap(ctx context.Context) (map[string][]db.Card, map[string]map[string]db.Staff, error) {
	var (
		mozuCards       = make(map[string][]db.Card)
		allCards        []db.Card
		mozuCardToStaff = make(map[string]map[string]db.Staff)
		err             error
	)
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if allCards, err = GetAllCards(tx, "", nil); err != nil {
			return err
		}

		for i := range allCards {
			c := &allCards[i]
			mozuCards[c.MozuID] = append(mozuCards[c.MozuID], *c)
		}

		for mozu, cards := range mozuCards {
			staffIDs := getCardsStaffID(cards)
			staffs, err := GetStaffsByID(tx, staffIDs)
			if err != nil {
				return err
			}

			cardToStaff := make(map[string]db.Staff, len(cards))
			for i := range cards {
				cardToStaff[cards[i].CardNo] = staffs[cards[i].StaffID]
			}

			mozuCardToStaff[mozu] = cardToStaff
		}
		return nil
	})

	return mozuCards, mozuCardToStaff, err
}

// updateGroupDoorAndCardRelation 更新权限组的门和卡关联关系
func updateGroupDoorAndCardRelation(tx *gorm.DB, id db.IDType, agWrapper db.AccessGroupInfoWrapper, mozuID string) error {
	// 添加权限与门的关系记录
	if agWrapper.Doors != nil {
		if err := updateAccessGroupDoorRelation(tx, id, agWrapper.Doors); err != nil {
			return err
		}
	}
	// 添加权限组关联的卡关系记录
	if agWrapper.Cards != nil {
		if err := updateAccessGroupCardsRelation(tx, id, agWrapper.Cards, mozuID); err != nil {
			return err
		}
	}

	return nil
}

// updateIndex 更新索引记录
func updateIndex(tx *gorm.DB, controllerID db.IDType, index, last int, value interface{}) error {
	return tgorm.WithOptions(tx.Model(value), withControllerIDOption(controllerID)).Updates(map[string]interface{}{
		db.ColumnIndex:      index,
		db.ColumnLast:       last,
		db.ColumnUpdateTime: time.Now().Unix(),
	}).Error
}

// updateCurrentSyncedTimestampIndexer 更新当前同步的时间戳索引
func updateCurrentSyncedTimestampIndexer(tx *gorm.DB,
	controllerID db.IDType, timestamp int64, value interface{},
) error {
	return tgorm.WithOptions(tx.Model(value), withControllerIDOption(controllerID)).Updates(map[string]interface{}{
		db.ColumnCurrentSyncedTimestamp: timestamp,
	}).Error
}

// updateHistorySyncedTimestampIndexer 更新历史同步的时间戳索引
func updateHistorySyncedTimestampIndexer(tx *gorm.DB,
	controllerID db.IDType, timestamp int64, value interface{},
) error {
	return tgorm.WithOptions(tx.Model(value), withControllerIDOption(controllerID)).Updates(map[string]interface{}{
		db.ColumnHistorySyncedTimestamp: timestamp,
	}).Error
}

// UpdateControllerGIDByCollectCode 根据采集编码更新控制器GID
func UpdateControllerGIDByCollectCode(tx *gorm.DB, code string, gid db.GIDType) error {
	if len(code) == 0 || len(gid) == 0 {
		return nil
	}

	return tgorm.WithOptions(tx.Model(&db.DoorController{}), withName(code)).Update(db.ColumnGID, gid).Error
}

// GetDoorCollectCodes 获取所有门的采集编码映射
func GetDoorCollectCodes(tx *gorm.DB) (map[string]db.IDType, error) {
	controllers, doors, err := GetAllDoorControllersAndDoors(tx, "")
	if err != nil {
		return nil, err
	}

	doorCodeIDMap := make(map[string]db.IDType)
	for i := range controllers {
		c := &controllers[i]

		ds, ok := doors[c.ID]
		if !ok {
			continue
		}

		controllerCode := c.GetCollectCode()

		for j := range ds {
			d := &ds[j]

			doorCodeIDMap[d.GetCollectCode(controllerCode)] = d.ID
		}
	}

	return doorCodeIDMap, nil
}

// UpdateDoorGIDByCollectCode 根据采集编码更新门GID
func UpdateDoorGIDByCollectCode(tx *gorm.DB, code string, gid db.GIDType, doorCodeIDMap map[string]db.IDType) error {
	if len(code) == 0 || len(gid) == 0 {
		return nil
	}

	var err error
	if doorCodeIDMap == nil {
		if doorCodeIDMap, err = GetDoorCollectCodes(tx); err != nil {
			return err
		}
	}

	if id, ok := doorCodeIDMap[code]; ok {
		return tgorm.WithOptions(tx.Model(&db.Door{}), withIDOption(id)).Update(db.ColumnGID, gid).Error
	}

	return nil
}

// updateControllerGIDsByCode 批量根据编码更新控制器GID
func updateControllerGIDsByCode(tx *gorm.DB, codeGIDs rt.CodeGIDMapType) error {
	var err error

	for code, gid := range codeGIDs {
		if err = UpdateControllerGIDByCollectCode(tx, code, gid); err != nil {
			return err
		}
	}

	return nil
}

// updateDoorGIDsByCode 批量根据编码更新门GID
func updateDoorGIDsByCode(tx *gorm.DB, codeGIDs rt.CodeGIDMapType) error {
	var err error

	doorCodes, err := GetDoorCollectCodes(tx)
	if err != nil {
		return err
	}

	for code, gid := range codeGIDs {
		if err = UpdateDoorGIDByCollectCode(tx, code, gid, doorCodes); err != nil {
			return err
		}
	}

	return nil
}

// Transaction 执行数据库事务
func (d *impl) Transaction(ctx context.Context, f func(tx *gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if f == nil {
			return nil
		}
		return f(tx)
	})
}
