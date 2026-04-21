// Package card 提供门禁卡的核心业务逻辑，包括卡与权限组的关联管理。
package card

import (
	"dac/entity/model/db"
)

// cancelDelete 取消对仍有门关联的控制器的删除操作
func cancelDelete(
	controllerDoors map[db.IDType]map[int]struct{},
	toDelete map[db.IDType]struct{},
) {
	var ok bool
	for controllerID, doors := range controllerDoors {
		if len(doors) == 0 {
			continue
		}
		if _, ok = toDelete[controllerID]; ok {
			delete(toDelete, controllerID)
		}
	}
}

// add 合并两个卡-控制器-门映射，计算需要新增和删除的条目
func add(
	lhs, rhs, toAdd map[string]map[db.IDType]map[int]struct{},
	toDelete map[string]map[db.IDType]struct{},
) (map[string]map[db.IDType]map[int]struct{},
	map[string]map[db.IDType]struct{}) {
	if toAdd == nil {
		toAdd = make(map[string]map[db.IDType]map[int]struct{})
	}
	if toDelete == nil {
		toDelete = make(map[string]map[db.IDType]struct{})
	}
	var ok bool
	for card, lhsControllerDoors := range lhs {
		if _, ok = toAdd[card]; !ok {
			toAdd[card] = make(map[db.IDType]map[int]struct{})
		}
		if _, ok = toDelete[card]; !ok {
			toDelete[card] = make(map[db.IDType]struct{})
		}

		cancelDelete(lhsControllerDoors, toDelete[card])

		rhsControllerDoors, ok := rhs[card]
		if !ok {
			toAdd[card] = lhsControllerDoors
		}

		for controllerID, lhsDoors := range lhsControllerDoors {
			if _, ok = toAdd[card][controllerID]; !ok {
				toAdd[card][controllerID] = make(map[int]struct{}, len(lhsDoors))
			}
			for lhsDoor := range lhsDoors {
				toAdd[card][controllerID][lhsDoor] = struct{}{}
			}
			for rhsDoor := range rhsControllerDoors[controllerID] {
				toAdd[card][controllerID][rhsDoor] = struct{}{}
			}
		}
	}

	return toAdd, toDelete
}

// AddCardControllerDoors 双向合并卡-控制器-门映射，返回需新增和删除的条目
func AddCardControllerDoors(
	lhs, rhs, toAdd map[string]map[db.IDType]map[int]struct{},
	toDelete map[string]map[db.IDType]struct{},
) (map[string]map[db.IDType]map[int]struct{},
	map[string]map[db.IDType]struct{}) {
	s1, s2 := add(lhs, rhs, toAdd, toDelete)
	return add(rhs, lhs, s1, s2)
}

// SubControllerDoors 计算两个卡-控制器-门映射的差集，返回需删除和需新增的条目
func SubControllerDoors(
	lhs, rhs map[string]map[db.IDType]map[int]struct{},
) (map[string]map[db.IDType]struct{},
	map[string]map[db.IDType]map[int]struct{}) {
	toDelete := make(map[string]map[db.IDType]struct{})
	toAdd := make(map[string]map[db.IDType]map[int]struct{})
	for card, lhsControllerDoors := range lhs {
		rhsControllerDoors, ok := rhs[card]
		if !ok {
			continue
		}

		toDelete[card] = make(map[db.IDType]struct{})
		toAdd[card] = make(map[db.IDType]map[int]struct{}, len(lhsControllerDoors))

		for controllerID, lhsDoors := range lhsControllerDoors {
			if !ok {
				continue
			}

			deleteDoorIDs := make([]int, len(lhsDoors))
			remainDoors := make(map[int]struct{})
			for door := range lhsDoors {
				if _, ok = rhsControllerDoors[door]; ok {
					deleteDoorIDs = append(deleteDoorIDs, door)
					continue
				}
				remainDoors[door] = struct{}{}
			}
			for _, door := range deleteDoorIDs {
				delete(lhsDoors, door)
			}

			if len(remainDoors) == 0 {
				// 若剩余门为空，则需要删除
				toDelete[card][controllerID] = struct{}{}
			} else {
				// 否则需要添加未被删除的门
				toAdd[card][controllerID] = remainDoors
			}
		}
	}
	return toDelete, toAdd
}
