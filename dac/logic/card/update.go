// Package card 提供门禁卡的增删改查业务逻辑，
// 包括卡与权限组、控制器、门的关联管理。
package card

import (
	"context"
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/cache"
	"dac/repo/dac"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// UpdateByAccessGroups 更新权限组时同步更新关联卡的控制器权限
func UpdateByAccessGroups(ctx context.Context, id db.IDType, mozuID string, wrapper db.AccessGroupInfoWrapper) error {
	var (
		err                               error
		accessGroupIDs                    = []db.IDType{id}
		newCards                          = wrapper.Cards
		newDoorIDs                        = wrapper.Doors
		oldCards                          []string
		timeGroup                         = wrapper.TimeGroupNo
		newCardControllerTimeGroups       = make(map[string]map[db.IDType]int)
		newCardControllerDoors            = make(map[string]map[db.IDType]map[int]struct{})
		oldCardControllerDoors            = make(map[string]map[db.IDType]map[int]struct{})
		oldFullCardControllerTimeGroups   = make(map[string]map[db.IDType]int)
		oldFullCardControllerDoors        = make(map[string]map[db.IDType]map[int]struct{})
		oldFullCardControllerAccessGroups = make(map[string]map[db.IDType]map[db.IDType]struct{})
		toDeleteCardControllers           = make(map[string]map[db.IDType]struct{})
		toAddCardControllerTimeGroups     = make(map[string]map[db.IDType]int)
		toAddCardControllerDoors          = make(map[string]map[db.IDType]map[int]struct{})
	)

	return dac.GetRW().UpdateAccessGroup(ctx, id, wrapper, mozuID, func(tx *gorm.DB) error {
		newDoors, err := dac.GetDoors(tx, newDoorIDs)
		if err != nil {
			return err
		}

		buildNewCardControllerMap(newCards, newDoors, timeGroup,
			newCardControllerTimeGroups, newCardControllerDoors)

		if _, oldCardControllerDoors, _, err = dac.GetCardCtrlTimeGroupDoorsByGroups(tx, accessGroupIDs); err != nil {
			return err
		}

		if oldCards, err = dac.GetCardNumbersByAccessGroupIDs(tx, accessGroupIDs); err != nil {
			return err
		}
		oldFullCardControllerTimeGroups, oldFullCardControllerDoors, oldFullCardControllerAccessGroups, err =
			dac.GetCardCtrlTimeGroupDoorsByCards(tx, oldCards, mozuID)
		if err != nil {
			return err
		}

		// 删除门禁卡在原有权限组的权限
		calcOldCardDeletions(oldCards, oldCardControllerDoors,
			oldFullCardControllerTimeGroups, oldFullCardControllerDoors,
			toDeleteCardControllers, toAddCardControllerDoors,
			toAddCardControllerTimeGroups)

		// 增加门禁卡在现门禁组的权限
		if err = calcNewCardAdditions(newCards, oldCardControllerDoors,
			oldFullCardControllerDoors, oldFullCardControllerTimeGroups,
			oldFullCardControllerAccessGroups,
			newCardControllerDoors, newCardControllerTimeGroups,
			toDeleteCardControllers, toAddCardControllerDoors,
			toAddCardControllerTimeGroups); err != nil {
			return err
		}

		return nil
	}, func(tx *gorm.DB) error {
		if err = DeleteInController(tx, toDeleteCardControllers); err != nil {
			return err
		}
		toAddCards := make([]string, 0, len(toAddCardControllerDoors))
		for card := range toAddCardControllerDoors {
			toAddCards = append(toAddCards, card)
		}
		if err = AddByControllerTimeGroupAndDoors(
			tx, toAddCards, mozuID,
			toAddCardControllerTimeGroups,
			toAddCardControllerDoors,
		); err != nil {
			return err
		}
		return nil
	})
}

// buildNewCardControllerMap 构建新卡号到控制器时间组和门的映射。
func buildNewCardControllerMap(
	newCards []string, newDoors []db.Door, timeGroup int,
	timeGroups map[string]map[db.IDType]int,
	doors map[string]map[db.IDType]map[int]struct{},
) {
	for _, card := range newCards {
		timeGroups[card] = make(map[db.IDType]int)
		doors[card] = make(map[db.IDType]map[int]struct{})

		for i := range newDoors {
			d := &newDoors[i]
			controllerID := d.ControllerID

			if _, ok := doors[card][controllerID]; !ok {
				doors[card][controllerID] = make(map[int]struct{})
			}
			doors[card][controllerID][d.Number] = struct{}{}
			timeGroups[card][controllerID] = timeGroup
		}
	}
}

// calcOldCardDeletions 计算旧卡号需要从哪些控制器删除或保留哪些门。
func calcOldCardDeletions(
	oldCards []string,
	oldCardControllerDoors map[string]map[db.IDType]map[int]struct{},
	oldFullTimeGroups map[string]map[db.IDType]int,
	oldFullDoors map[string]map[db.IDType]map[int]struct{},
	toDelete map[string]map[db.IDType]struct{},
	toAddDoors map[string]map[db.IDType]map[int]struct{},
	toAddTimeGroups map[string]map[db.IDType]int,
) {
	for _, oldCard := range oldCards {
		toDelete[oldCard] = make(map[db.IDType]struct{})
		toAddDoors[oldCard] = make(map[db.IDType]map[int]struct{})
		toAddTimeGroups[oldCard] = oldFullTimeGroups[oldCard]

		oldControllerDoors, ok := oldCardControllerDoors[oldCard]
		if !ok {
			continue
		}

		oldFullControllerDoors := oldFullDoors[oldCard]
		for controllerID, fullDoors := range oldFullControllerDoors {
			doors, ok := oldControllerDoors[controllerID]
			if !ok {
				continue
			}
			if len(doors) == len(fullDoors) {
				toDelete[oldCard][controllerID] = struct{}{}
			} else {
				toAddDoors[oldCard][controllerID] = make(map[int]struct{})
				for d := range fullDoors {
					if _, ok = doors[d]; ok {
						continue
					}
					toAddDoors[oldCard][controllerID][d] = struct{}{}
				}
			}
		}
	}
}

// calcNewCardAdditions 计算新卡号需要添加到哪些控制器及对应的门和时间组。
func calcNewCardAdditions(
	newCards []string,
	oldCardControllerDoors map[string]map[db.IDType]map[int]struct{},
	oldFullDoors map[string]map[db.IDType]map[int]struct{},
	oldFullTimeGroups map[string]map[db.IDType]int,
	oldFullAccessGroups map[string]map[db.IDType]map[db.IDType]struct{},
	newControllerDoorsMap map[string]map[db.IDType]map[int]struct{},
	newControllerTimeGroupsMap map[string]map[db.IDType]int,
	toDelete map[string]map[db.IDType]struct{},
	toAddDoors map[string]map[db.IDType]map[int]struct{},
	toAddTimeGroups map[string]map[db.IDType]int,
) error {
	for _, newCard := range newCards {
		ensureCardMaps(newCard, toDelete, toAddDoors, toAddTimeGroups)

		oldControllerDoors := oldCardControllerDoors[newCard]

		oldFullControllerDoors, ok := oldFullDoors[newCard]
		if !ok {
			toAddDoors[newCard] = newControllerDoorsMap[newCard]
			toAddTimeGroups[newCard] = newControllerTimeGroupsMap[newCard]
			continue
		}
		fullControllerTimeGroups := oldFullTimeGroups[newCard]

		newControllerDoors := newControllerDoorsMap[newCard]
		newControllerTimeGroups := newControllerTimeGroupsMap[newCard]
		for controllerID, newDoors := range newControllerDoors {
			if err := mergeControllerDoors(
				newCard, controllerID, newDoors,
				oldControllerDoors, oldFullControllerDoors,
				fullControllerTimeGroups, newControllerTimeGroups,
				oldFullAccessGroups,
				toDelete, toAddDoors, toAddTimeGroups,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureCardMaps 确保卡号在各映射中已初始化。
func ensureCardMaps(
	card string,
	toDelete map[string]map[db.IDType]struct{},
	toAddDoors map[string]map[db.IDType]map[int]struct{},
	toAddTimeGroups map[string]map[db.IDType]int,
) {
	if _, ok := toDelete[card]; !ok {
		toDelete[card] = make(map[db.IDType]struct{})
	}
	if _, ok := toAddDoors[card]; !ok {
		toAddDoors[card] = make(map[db.IDType]map[int]struct{})
	}
	if _, ok := toAddTimeGroups[card]; !ok {
		toAddTimeGroups[card] = make(map[db.IDType]int)
	}
}

// mergeControllerDoors 合并单个控制器上新旧门的权限变更。
func mergeControllerDoors(
	card string, controllerID db.IDType,
	newDoors map[int]struct{},
	oldControllerDoors map[db.IDType]map[int]struct{},
	oldFullControllerDoors map[db.IDType]map[int]struct{},
	fullTimeGroups map[db.IDType]int,
	newTimeGroups map[db.IDType]int,
	oldFullAccessGroups map[string]map[db.IDType]map[db.IDType]struct{},
	toDelete map[string]map[db.IDType]struct{},
	toAddDoors map[string]map[db.IDType]map[int]struct{},
	toAddTimeGroups map[string]map[db.IDType]int,
) error {
	if _, ok := toAddDoors[card][controllerID]; !ok {
		toAddDoors[card][controllerID] = make(map[int]struct{})
	}

	canResetTimeGroup := false
	if _, ok1 := toDelete[card]; ok1 {
		if _, ok2 := toDelete[card][controllerID]; ok2 {
			delete(toDelete[card], controllerID)
			canResetTimeGroup = len(oldFullAccessGroups[card][controllerID]) < 2
		}
	}

	oldTimeGroup, ok1 := fullTimeGroups[controllerID]
	newTimeGroup, ok2 := newTimeGroups[controllerID]
	if ok1 && ok2 && !canResetTimeGroup && oldTimeGroup != newTimeGroup {
		return fmt.Errorf("门禁卡: %v, 门禁控制器: %v，时间组不一致: %v, %v",
			card, controllerID, oldTimeGroup, newTimeGroup)
	}
	toAddTimeGroups[card][controllerID] = newTimeGroup

	for d := range newDoors {
		toAddDoors[card][controllerID][d] = struct{}{}
	}

	oldDoors := oldControllerDoors[controllerID]
	for d := range oldFullControllerDoors[controllerID] {
		if _, ok := oldDoors[d]; ok {
			continue
		}
		toAddDoors[card][controllerID][d] = struct{}{}
	}
	return nil
}

// UpdateToAccessGroups 将卡号批量关联到指定权限组，
// 同步计算需要在控制器上增删的卡权限。
func UpdateToAccessGroups(ctx context.Context, cardNumbers []string, accessGroupIDs []int, mozuID string) error {
	var (
		err                           error
		currentControllerTimeGroups   = make(map[db.IDType]int)                         // 当前权限组的控制器时间组
		currentControllerDoors        = make(map[db.IDType]map[int]struct{})            // 当前权限组的控制器门
		toDeleteCardControllers       = make(map[string]map[db.IDType]struct{})         // 需要从中删除的门禁控制器
		toAddCardControllerDoors      = make(map[string]map[db.IDType]map[int]struct{}) // 需要添加或设置的门禁控制器及门
		toAddCardControllerTimeGroups = make(map[string]map[db.IDType]int)              // 需要添加的控制器时间组
	)
	return dac.GetRW().UpdateCardAccessGroupRelation(ctx, cardNumbers, accessGroupIDs, mozuID, true, func(tx *gorm.DB) error {
		currentControllerTimeGroups, currentControllerDoors, err = dac.GetControllerTimeGroupAndDoors(tx, accessGroupIDs)
		if err != nil {
			return err
		}

		oldCardControllerTimeGroups, oldCardControllerDoors, _, err :=
			dac.GetCardCtrlTimeGroupDoorsByCards(tx, cardNumbers, mozuID)
		if err != nil {
			return err
		}

		for card, oldControllerDoors := range oldCardControllerDoors {
			toDeleteCardControllers[card] = make(map[db.IDType]struct{})
			toAddCardControllerDoors[card] = make(map[db.IDType]map[int]struct{})
			toAddCardControllerTimeGroups[card] = make(map[db.IDType]int)

			oldControllerTimeGroups := oldCardControllerTimeGroups[card]
			for controllerID, oldDoors := range oldControllerDoors {
				if len(oldDoors) == 0 {
					continue
				}

				// 若先前权限组有该控制器，但当前权限组无，
				// 则需要将卡从该控制器删除
				currentDoors, ok := currentControllerDoors[controllerID]
				if !ok || len(currentDoors) == 0 {
					toDeleteCardControllers[card][controllerID] = struct{}{}
					continue
				}

				// 此处先前权限组与当前权限组均有门
				currentTimeGroup, ok := currentControllerTimeGroups[controllerID]
				if ok && oldControllerTimeGroups[controllerID] != currentTimeGroup && !isDoorsBelong(oldDoors, currentDoors) {
					// 若先前门未被当前门包含，需要保证时间组一致
					return fmt.Errorf("门禁卡 \"%v\", 门禁控制器 %v 时间组不一致: %v, %v",
						card, controllerID, currentTimeGroup, oldControllerTimeGroups[controllerID])
				}
				if !ok {
					currentTimeGroup = oldControllerTimeGroups[controllerID]
				}

				toAddCardControllerTimeGroups[card][controllerID] = currentTimeGroup

				// 否则需要设置当前权限组在该控制器下可访问的门
				if _, ok := toAddCardControllerDoors[card][controllerID]; !ok {
					toAddCardControllerDoors[card][controllerID] = make(map[int]struct{})
				}
				addDoorMap(toAddCardControllerDoors[card][controllerID], currentDoors)
			}

			for controllerID, currentDoors := range currentControllerDoors {
				if _, ok := oldControllerDoors[controllerID]; ok {
					continue
				}
				// 若先前权限组不存在该控制器，则需要添加
				toAddCardControllerDoors[card][controllerID] = currentDoors
				toAddCardControllerTimeGroups[card][controllerID] = currentControllerTimeGroups[controllerID]
			}
		}

		for _, card := range cardNumbers {
			if _, ok := oldCardControllerDoors[card]; ok {
				continue
			}

			// 未在先前权限组的控制器，均需要添加
			toAddCardControllerDoors[card] = currentControllerDoors
			toAddCardControllerTimeGroups[card] = currentControllerTimeGroups
		}

		return nil
	}, func(tx *gorm.DB) error {
		if err = DeleteInController(tx, toDeleteCardControllers); err != nil {
			return err
		}

		if err = AddByControllerTimeGroupAndDoors(
			tx, cardNumbers, mozuID,
			toAddCardControllerTimeGroups,
			toAddCardControllerDoors,
		); err != nil {
			return err
		}
		return nil
	})
}

// UpdateStaff 更新卡关联的人员信息
func UpdateStaff(ctx context.Context, cards []string, staff int, mozuID string) error {
	return dac.GetRW().UpdateCardsStaff(ctx, cards, staff, mozuID, func(tx *gorm.DB) error {
		return updateStaff(tx, cards, staff, mozuID)
	})
}

// UpdateStaffInController 向控制器下发更新人员信息的请求
func UpdateStaffInController(tx *gorm.DB, cardNumbers []string, mozuID, username, password string, staffID int) error {
	if len(cardNumbers) == 0 {
		return nil
	}

	payload := driver.Card{
		UserName: username,
		Password: password,
	}

	cardControllerMap, err := dac.GetControllerIDsByCards(tx, cardNumbers, mozuID)
	if err != nil {
		return err
	}

	reqs := make([]db.Request, 0, len(cardNumbers))
	for card, controllerIDMap := range cardControllerMap {
		p := payload
		p.CardNo = card

		b, err := driver.Marshal(p)
		if err != nil {
			return err
		}

		for controllerID := range controllerIDMap {
			controller, _ := cache.Get().GetController(controllerID)

			// CACS 门控器的 UserName 只支持数字，使用 StaffID
			cp := p
			if controller.Protocol.Name == consts.ProtocolCACS {
				cp.UserName = strconv.Itoa(staffID)
			}

			b, err = driver.Marshal(cp)
			if err != nil {
				return err
			}

			req := db.Request{
				ControllerID: controllerID,
				Method:       driver.MethodUpdateCardStaff,
				Payload:      b,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
			reqs = append(reqs, req)
		}
	}

	return dac.AddRequests(tx, reqs)
}

// updateStaff 根据人员ID查询人员信息并下发到控制器
func updateStaff(tx *gorm.DB, cardNumbers []string, staffID int, mozuID string) error {
	staff, err := dac.GetStaffByID(tx, staffID)
	if err != nil {
		return fmt.Errorf("get staff %v error: %w", staffID, err)
	}

	return UpdateStaffInController(tx, cardNumbers, mozuID, staff.Name, staff.Password, staff.ID)
}

// UpdateFlagInTransaction 在事务中更新卡标志（启用/禁用）
func UpdateFlagInTransaction(tx *gorm.DB, cards []string, mozuID string, flag db.CardFlagType) error {
	return dac.UpdateCardsFlag(tx, cards, mozuID, flag, func(tx *gorm.DB) error {
		return updateFlagInController(tx, cards, mozuID, flag)
	})
}

// UpdateFlag 更新卡标志并同步到控制器
func UpdateFlag(ctx context.Context, cards []string, mozuID string, flag db.CardFlagType) error {
	return dac.GetRW().UpdateCardsFlag(ctx, cards, mozuID, flag, func(tx *gorm.DB) error {
		return updateFlagInController(tx, cards, mozuID, flag)
	})
}

// UpdateType 更新卡类型
func UpdateType(ctx context.Context, cards []string, mozuID string, cardType db.CardType) error {
	return dac.GetRW().UpdateCardsType(ctx, cards, mozuID, cardType)
}

// UpdateValidTime 更新卡有效期并同步到控制器
func UpdateValidTime(ctx context.Context, cards []string, mozuID string, validTime int64) error {
	return dac.GetRW().UpdateCardValidTime(ctx, cards, mozuID, validTime, func(tx *gorm.DB) error {
		return updateValidInController(tx, cards, mozuID, db.CardFlagEnable)
	})
}

// updateFlagInController 向控制器下发更新卡标志的请求
func updateFlagInController(tx *gorm.DB, cards []string, mozuID string, flag db.CardFlagType) error {
	if len(cards) == 0 {
		return fmt.Errorf("待更新有效期的卡号为空")
	}

	cardControllers, err := dac.GetControllerIDsByCards(tx, cards, mozuID)
	if err != nil {
		return err
	}

	return buildAndAddCardFlagRequests(tx, cardControllers, flag)
}

// updateValidInController 向控制器下发续期卡的启用请求（仅处理非禁用卡）
func updateValidInController(tx *gorm.DB, cards []string, mozuID string, flag db.CardFlagType) error {
	if len(cards) == 0 {
		return fmt.Errorf("待更新有效期的卡号为空")
	}

	// 筛选只因过期而无效，且此次是续期的卡
	dbCards, err := dac.GetCardsByCardNos(tx, cards, mozuID)
	toBeUpdated := make([]string, 0, len(dbCards))
	if len(dbCards) == 0 {
		return fmt.Errorf("数据库中不存在卡：%v", dbCards)
	}
	for _, dbCard := range dbCards {
		if dbCard.CardFlag != db.CardFlagDisable {
			toBeUpdated = append(toBeUpdated, dbCard.CardNo)
		}
	}

	cardControllers, err := dac.GetControllerIDsByCards(tx, toBeUpdated, mozuID)
	if err != nil {
		return err
	}

	return buildAndAddCardFlagRequests(tx, cardControllers, flag)
}

// buildAndAddCardFlagRequests 根据卡号与门控器的映射关系，构造更新卡片标志的请求并批量添加
func buildAndAddCardFlagRequests(tx *gorm.DB, cardControllers map[string]map[db.IDType]struct{}, flag db.CardFlagType) error {
	reqs := make([]db.Request, 0, len(cardControllers))
	for card, controllers := range cardControllers {
		payload := driver.Card{
			CardNo:   card,
			CardFlag: int(flag),
		}

		b, err := driver.Marshal(payload)
		if err != nil {
			continue
		}

		// 构造请求
		for controllerID := range controllers {
			controller, _ := cache.Get().GetController(controllerID)
			req := db.Request{
				ControllerID: controllerID,
				Method:       driver.MethodUpdateCardFlag,
				Payload:      b,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
			reqs = append(reqs, req)
		}
	}

	return dac.AddRequests(tx, reqs)
}
