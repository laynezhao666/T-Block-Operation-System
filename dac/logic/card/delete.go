// Package card 实现门禁卡的增删改查和权限管理功能。
package card

import (
	"context"
	"dac/entity/consts"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/cache"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// DeleteInTransaction 在指定事务中删除门禁卡，同时清理控制器中的卡数据
func DeleteInTransaction(tx *gorm.DB, cards []string, mozuID string) error {
	return dac.DeleteCards(tx, cards, mozuID, func(tx *gorm.DB) error {
		return deleteInController(tx, cards, mozuID)
	})
}

// Delete 删除门禁卡，自动管理事务
func Delete(ctx context.Context, cards []string, mozuID string) error {
	return dac.GetRW().DeleteCards(ctx, cards, mozuID, func(tx *gorm.DB) error {
		return deleteInController(tx, cards, mozuID)
	})
}

// DeleteInController 根据卡号和控制器ID映射，向指定控制器下发删卡请求
func DeleteInController(tx *gorm.DB, cardControllerIDs map[string]map[db.IDType]struct{}) error {
	if len(cardControllerIDs) == 0 {
		return nil
	}

	reqs := make([]db.Request, 0, len(cardControllerIDs))
	for card, controllerIDs := range cardControllerIDs {
		b, err := driver.Marshal(card)
		if err != nil {
			return err
		}

		for controllerID := range controllerIDs {
			controller, _ := cache.Get().GetController(controllerID)
			req := db.Request{
				ControllerID: controllerID,
				Method:       driver.MethodDeleteCard,
				Payload:      b,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
			reqs = append(reqs, req)
		}
	}

	return dac.AddRequests(tx, reqs)
}

// DeleteInControllerAll 向所有指定控制器下发删卡请求。
// 对于V3协议控制器使用DeleteUser方法，其他协议使用DeleteCard方法。
func DeleteInControllerAll(
	tx *gorm.DB, cards []string,
	controllerIDs []db.IDType, staff map[string]db.Staff,
) error {
	if len(cards) == 0 || len(controllerIDs) == 0 {
		return nil
	}

	reqs := make([]db.Request, 0, len(controllerIDs)*len(cards))
	for i := range controllerIDs {
		controller, _ := cache.Get().GetController(controllerIDs[i])
		// 判断是否为V3协议
		isV3 := controller.Protocol.Name == consts.ProtocolHTTP &&
			controller.Protocol.Version == consts.V3ProtocolVersion

		var req db.Request
		if isV3 {
			// V3协议使用DeleteUser方法
			req = db.Request{
				ControllerID: controllerIDs[i],
				Method:       driver.MethodDeleteUser,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
			for c := range cards {
				user := driver.UserID{
					UserID: staff[cards[c]].ID,
				}
				b, err := driver.Marshal(user)
				if err != nil {
					return err
				}
				req.Payload = b
				reqs = append(reqs, req)
			}
		} else {
			// 非V3协议使用DeleteCard方法
			req = db.Request{
				ControllerID: controllerIDs[i],
				Method:       driver.MethodDeleteCard,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
			for j := range cards {
				b, err := driver.Marshal(cards[j])
				if err != nil {
					return err
				}
				req.Payload = b
				reqs = append(reqs, req)
			}
		}

	}

	return dac.AddRequests(tx, reqs)
}

// deleteByAccessGroupInController 根据权限组ID查找关联控制器，然后删除卡
func deleteByAccessGroupInController(
	tx *gorm.DB, cards []string,
	accessGroupIDs []db.IDType, staff map[string]db.Staff,
) error {
	controllerIDs, err := dac.GetControllerIDsByAccessGroupIDs(
		tx, accessGroupIDs)
	if err != nil {
		return err
	}

	return DeleteInControllerAll(tx, cards, controllerIDs, staff)
}

// deleteInController 删除卡时，查找关联的权限组和控制器，下发删卡请求
func deleteInController(tx *gorm.DB, cards []string, mozuID string) error {
	if len(cards) == 0 {
		return nil
	}

	// 获取卡关联的权限组
	accessGroupIDs, err := dac.GetAccessGroupIDByCards(tx, cards, mozuID)
	if err != nil {
		return err
	}
	if len(accessGroupIDs) == 0 {
		return nil
	}

	// 获取卡关联的员工信息
	_, staffs, err := dac.GetCardStaffMapByCards(tx, cards, mozuID)
	if err != nil {
		return err
	}

	return deleteByAccessGroupInController(tx, cards, accessGroupIDs, staffs)
}

// PruneAccessGroupInController 删除权限组时，清理控制器中的关联卡数据。
// 对于部分门被删除的卡，更新其在控制器中的门编号；
// 对于所有门都被删除的卡，直接从控制器中删除。
func PruneAccessGroupInController(ctx context.Context, id db.IDType, mozuID string) error {
	var (
		cards                         []string
		err                           error
		accessGroupIDs                = []db.IDType{id}
		toDeleteCardControllerDoors   map[string]map[db.IDType]map[int]struct{}
		oldCardControllerDoors        map[string]map[db.IDType]map[int]struct{}
		oldCardControllerTimeGroups   map[string]map[db.IDType]int
		toDeleteCardControllerIDs     = make(map[string]map[db.IDType]struct{})
		toAddCardControllerDoors      = make(map[string]map[db.IDType]map[int]struct{})
		toAddCardControllerTimeGroups = make(map[string]map[db.IDType]int)
	)
	return dac.GetRW().DeleteAccessGroup(ctx, id, func(tx *gorm.DB) error {
		_, toDeleteCardControllerDoors, _, err = dac.GetCardCtrlTimeGroupDoorsByGroups(tx, accessGroupIDs)
		if err != nil {
			return err
		}

		if cards, err = dac.GetCardNumbersByAccessGroupIDs(tx, accessGroupIDs); err != nil {
			return err
		}
		oldCardControllerTimeGroups, oldCardControllerDoors, _, err =
			dac.GetCardCtrlTimeGroupDoorsByCards(tx, cards, mozuID)
		if err != nil {
			return err
		}

		for card, oldControllerDoors := range oldCardControllerDoors {
			toDeleteCardControllerIDs[card] = make(map[db.IDType]struct{})
			toAddCardControllerDoors[card] = make(map[db.IDType]map[int]struct{})
			toAddCardControllerTimeGroups[card] = make(map[db.IDType]int)

			toDeleteControllerDoors, ok := toDeleteCardControllerDoors[card]
			if !ok {
				continue
			}

			for controllerID, oldDoors := range oldControllerDoors {
				toAddCardControllerDoors[card][controllerID] = make(map[int]struct{})

				toDeleteDoors, ok := toDeleteControllerDoors[controllerID]
				if !ok {
					// 门禁卡在该控制器下不需要删除，跳过
					continue
				}

				if len(toDeleteDoors) == len(oldDoors) {
					// 若门禁卡在该控制器下的所有门都需要删除
					toDeleteCardControllerIDs[card][controllerID] = struct{}{}
					continue
				}

				// 否则需要更新门禁卡在该控制器的门编号
				for door := range oldDoors {
					if _, ok = toDeleteDoors[door]; ok {
						continue
					}
					toAddCardControllerDoors[card][controllerID][door] = struct{}{}
				}

				toAddCardControllerTimeGroups[card][controllerID] = oldCardControllerTimeGroups[card][controllerID]
			}
		}
		return nil
	}, func(tx *gorm.DB) error {
		if err = DeleteInController(tx, toDeleteCardControllerIDs); err != nil {
			return err
		}
		if err = AddByControllerTimeGroupAndDoors(
			tx, cards, mozuID,
			toAddCardControllerTimeGroups,
			toAddCardControllerDoors,
		); err != nil {
			return err
		}
		return nil
	})
}
