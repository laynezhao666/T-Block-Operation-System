package card

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/cache"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// AddCard 添加门禁卡，绑定权限组关系并下发到门禁控制器。
func AddCard(
	ctx context.Context, mozuID string,
	cardNumber string, cardFlag, cardType int,
	validTime int64, staffID int,
	accessGroupIDs []int,
) error {
	card := db.Card{
		CardNo:    cardNumber,
		CardFlag:  db.CardFlagType(cardFlag),
		CardType:  db.CardType(cardType),
		ValidTime: validTime,
		StaffID:   staffID,
		MozuID:    mozuID,
	}

	var err error

	// 事务执行：1、绑定卡和权限组的关系 2、在门禁控制器上加卡
	return dac.GetRW().AddCard(ctx, card, func(tx *gorm.DB) error {
		if err = dac.AddCardsAccessGroupRelation(tx, []string{cardNumber}, accessGroupIDs, mozuID); err != nil {
			return err
		}
		if err = AddByStaffInControllers(tx, cardNumber, cardFlag, staffID, accessGroupIDs); err != nil {
			return err
		}
		return nil
	})
}

// AddByControllerTimeGroupAndDoors 根据控制器时间组和门列表批量添加卡。
// 查询卡关联的员工信息，构建请求下发到各门控器。
func AddByControllerTimeGroupAndDoors(
	tx *gorm.DB, cardNumbers []string, mozuID string,
	cardControllerTimeGroups map[string]map[db.IDType]int,
	cardControllerDoors map[string]map[db.IDType]map[int]struct{},
) error {
	cards, cardStaffMap, err := dac.GetCardStaffMapByCards(tx, cardNumbers, mozuID)
	if err != nil {
		return err
	}

	type account struct {
		Username string
		Password string
	}
	accounts := make(map[string]account)

	for i := range cards {
		cardNumber := cards[i].CardNo
		staff, ok := cardStaffMap[cardNumber]
		if !ok {
			accounts[cardNumber] = account{
				Username: consts.DefaultUserName,
				Password: consts.DefaultPassword,
			}
		} else {
			accounts[cardNumber] = account{
				Username: staff.Name,
				Password: staff.Password,
			}
		}
	}

	for i := range cards {
		c := &cards[i]
		cardNumber := cards[i].CardNo
		acc := accounts[cardNumber]
		if err = AddInController(
			tx, cardNumber, int(c.CardFlag),
			cardStaffMap[cardNumber], acc.Username, acc.Password,
			cardControllerTimeGroups[cardNumber],
			cardControllerDoors[c.CardNo],
		); err != nil {
			return err
		}
	}

	return nil
}

// addDoors 将门编号列表添加到门集合中。
func addDoors(doorMap map[int]struct{}, doors []int) {
	for _, d := range doors {
		doorMap[d] = struct{}{}
	}
}

// isDoorsBelong lhs 是否被 rhs 包含
func isDoorsBelong(lhs, rhs map[int]struct{}) bool {
	ok := false
	for k := range lhs {
		if _, ok = rhs[k]; !ok {
			return false
		}
	}
	return true
}

// addDoorMap 将门集合合并到目标门集合中。
func addDoorMap(doorMap map[int]struct{}, doors map[int]struct{}) {
	for d := range doors {
		doorMap[d] = struct{}{}
	}
}

// getDoorsFromMap 将门集合转换为门编号切片。
func getDoorsFromMap(doorMap map[int]struct{}) []int {
	doors := make([]int, 0, len(doorMap))
	for d := range doorMap {
		doors = append(doors, d)
	}
	return doors
}

// MergeByDoorsInController 合并权限组的门列表并下发到门控器。
// 将新权限组的门与已有权限组的门合并，确保时间组一致性。
func MergeByDoorsInController(
	ctx context.Context, mozuID string,
	wrapper db.AccessGroupInfoWrapper,
) (db.IDType, error) {
	var (
		err                           error
		oldCardControllerTimeGroups   map[string]map[db.IDType]int
		oldCardControllerDoors        map[string]map[db.IDType]map[int]struct{}
		doorIDs                       = wrapper.Doors
		cardNumbers                   = wrapper.Cards
		doors                         []db.Door
		timeGroupNumber               = wrapper.TimeGroupNo
		finalCardControllerDoors      = make(map[string]map[db.IDType]map[int]struct{}, len(oldCardControllerDoors))
		finalCardControllerTimeGroups = make(map[string]map[db.IDType]int, len(oldCardControllerTimeGroups))
	)
	wrapper.MozuID = mozuID

	return dac.GetRW().AddAccessGroup(ctx, wrapper, mozuID, func(tx *gorm.DB) error {
		oldCardControllerTimeGroups, oldCardControllerDoors, _, err =
			dac.GetCardCtrlTimeGroupDoorsByCards(tx, cardNumbers, mozuID)
		if err != nil {
			return err
		}

		doors, err = dac.GetDoors(tx, doorIDs)
		if err != nil {
			return err
		}

		currentControllerDoors := make(map[db.IDType][]int, len(doors))
		currentControllerTimeGroups := make(map[db.IDType]int, len(doors))
		for i := range doors {
			d := &doors[i]
			currentControllerDoors[d.ControllerID] = append(currentControllerDoors[d.ControllerID], d.Number)
			currentControllerTimeGroups[d.ControllerID] = timeGroupNumber
		}

		finalCardControllerDoors = make(map[string]map[db.IDType]map[int]struct{}, len(oldCardControllerDoors))
		finalCardControllerTimeGroups = make(map[string]map[db.IDType]int, len(oldCardControllerTimeGroups))

		// 需要合并先前权限组与当前权限门的时间组和门
		for card, oldControllerDoors := range oldCardControllerDoors {
			finalCardControllerDoors[card] = make(map[db.IDType]map[int]struct{})
			finalCardControllerTimeGroups[card] = make(map[db.IDType]int)
			oldControllerTimeGroups := oldCardControllerTimeGroups[card]

			for controllerID, currentDoors := range currentControllerDoors {
				currentTimeGroup := currentControllerTimeGroups[controllerID]
				if oldTimeGroup, ok := oldControllerTimeGroups[controllerID]; ok {
					if oldTimeGroup != currentTimeGroup {
						return fmt.Errorf(
							"门禁卡 \"%v\" 在门禁控制器 %v "+
								"的先前时间组号: %v 与当前时间组号: %v 不一致，"+
								"请检查时间组设置",
							card, controllerID,
							oldTimeGroup, currentTimeGroup)
					}
				}

				if _, ok := finalCardControllerDoors[card][controllerID]; !ok {
					finalCardControllerDoors[card][controllerID] = make(map[int]struct{})
				}
				addDoorMap(finalCardControllerDoors[card][controllerID], oldControllerDoors[controllerID])
				addDoors(finalCardControllerDoors[card][controllerID], currentDoors)

				finalCardControllerTimeGroups[card][controllerID] = currentTimeGroup
			}

			for controllerID, oldDoors := range oldControllerDoors {
				if _, ok := currentControllerDoors[controllerID]; ok {
					continue
				}

				if _, ok := finalCardControllerDoors[card][controllerID]; !ok {
					finalCardControllerDoors[card][controllerID] = make(map[int]struct{})
				}
				addDoorMap(finalCardControllerDoors[card][controllerID], oldDoors)
				finalCardControllerTimeGroups[card][controllerID] = oldControllerTimeGroups[controllerID]
			}
		}

		return nil
	}, func(tx *gorm.DB) error {
		return AddByControllerTimeGroupAndDoors(
			tx, cardNumbers, mozuID,
			finalCardControllerTimeGroups,
			finalCardControllerDoors)
	})
}

// AddInController 将卡信息下发到指定门控器。
// 根据协议类型（HTTP V3/CACS等）构建不同的请求格式。
func AddInController(
	tx *gorm.DB, cardNumber string,
	cardFlag int, staff db.Staff,
	username, password string,
	controllerTimeGroups map[db.IDType]int,
	controllerDoors map[db.IDType]map[int]struct{},
) error {
	if len(controllerDoors) == 0 {
		return nil
	}

	reqs := make([]db.Request, 0, len(controllerTimeGroups))
	for controllerID, timeGroupNumber := range controllerTimeGroups {
		controller, _ := cache.Get().GetController(controllerID)
		isV3 := controller.Protocol.Name == consts.ProtocolHTTP &&
			controller.Protocol.Version == consts.V3ProtocolVersion

		// 必须忽略门列表为空的情况
		doorNos := getDoorsFromMap(controllerDoors[controllerID])
		if len(doorNos) == 0 {
			continue
		}

		var req db.Request
		if isV3 {
			// 如果有人脸照片信息，base64编码，适用于自研门禁，北向http v3版本协议
			if staff.Picture != "" {
				// 删除 "base64," 及其之前的部分
				if idx := strings.Index(staff.Picture, "base64,"); idx != -1 {
					staff.Picture = staff.Picture[idx+len("base64,"):]
				}
				pic := staff.Picture
				prefix := pic
				suffix := pic
				if len(pic) > 120 {
					prefix = pic[:80]
					suffix = pic[len(pic)-40:]
				}
				config.Log.Infof(
					"[AddInController] v3 FaceImage 处理后长度=%d, "+
						"plus=%d, eq=%d, prefix=%s, suffix=%s",
					len(pic), strings.Count(pic, "+"),
					strings.Count(pic, "="), prefix, suffix)
				_, err := base64.StdEncoding.DecodeString(pic)
				if err != nil {
					config.Log.Errorf(
						"[AddInController] v3 FaceImage base64 校验失败: %v, len=%d",
						err, len(pic))
				} else {
					config.Log.Infof(
						"[AddInController] v3 FaceImage base64 校验通过, len=%d",
						len(pic))
				}
			}
			user := driver.CardWithStaffInfo{
				Card: driver.Card{
					CardNo:      cardNumber,
					CardFlag:    cardFlag,
					UserName:    username,
					Password:    password,
					TimeGroupNo: timeGroupNumber,
					DoorNos:     getDoorsFromMap(controllerDoors[controllerID]),
				},
				UserID:      staff.ID,
				FaceImage:   staff.Picture,
				FingerPrint: staff.Fingerprint,
			}
			b, err := driver.Marshal(user)
			if err != nil {
				continue
			}

			req = db.Request{
				ControllerID: controllerID,
				Method:       driver.MethodAddUser,
				Payload:      b,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
		} else {
			// CACS 门控器的 UserName 只支持数字，使用 StaffID
			cardUserName := username
			if controller.Protocol.Name == consts.ProtocolCACS {
				cardUserName = strconv.Itoa(staff.ID)
			}

			card := driver.Card{
				CardNo:      cardNumber,
				CardFlag:    cardFlag,
				UserName:    cardUserName,
				Password:    password,
				TimeGroupNo: timeGroupNumber,
				DoorNos:     getDoorsFromMap(controllerDoors[controllerID]),
			}
			b, err := driver.Marshal(card)
			if err != nil {
				continue
			}
			req = db.Request{
				ControllerID: controllerID,
				Method:       driver.MethodAddCard,
				Payload:      b,
				MozuID:       controller.MozuID,
				State:        consts.StateToBeExecuted,
			}
		}

		reqs = append(reqs, req)
	}

	return dac.AddRequests(tx, reqs)
}

// AddByAccessGroupInController 根据权限组ID列表将卡下发到门控器。
func AddByAccessGroupInController(
	tx *gorm.DB, cardNumber string,
	cardFlag int, staff db.Staff,
	username, password string,
	accessGroupIDs []db.IDType,
) error {
	controllerTimeGroups, controllerDoorNumbers, err :=
		dac.GetControllerTimeGroupAndDoors(tx, accessGroupIDs)
	if err != nil {
		return err
	}

	return AddInController(
		tx, cardNumber, cardFlag, staff,
		username, password,
		controllerTimeGroups, controllerDoorNumbers)
}

// AddByStaffInControllers 获取staff用户名密码，下发到门禁控制器。
func AddByStaffInControllers(
	tx *gorm.DB, cardNumber string,
	cardFlag int, staffID db.IDType,
	accessGroupIDs []db.IDType,
) error {
	username := ""
	password := ""

	staffs, err := dac.GetStaffsByID(tx, []db.IDType{staffID})
	if err != nil {
		return err
	}
	staff, ok := staffs[staffID]
	if !ok {
		username = consts.DefaultUserName
		password = consts.DefaultPassword
	} else {
		username = staff.Name
		password = staff.Password
	}

	return AddByAccessGroupInController(tx, cardNumber, cardFlag, staff, username, password, accessGroupIDs)
}
