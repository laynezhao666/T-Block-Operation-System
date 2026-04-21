// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"

	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/repo/dac"
)

// GetCards 分页查询门禁卡列表，支持按卡号、类型、状态、权限组等条件过滤。
// 返回总数、卡列表（含人员和权限组信息）和错误。
// 查询流程：先获取卡记录，再关联人员信息和权限组信息，最后组装为CGI响应格式。
func GetCards(ctx context.Context, mozuID string,
	offset, limit int, cards []string,
	query string, cardType db.CardType, queryCardType bool,
	cardFlag db.CardFlagType, queryCardFlag bool,
	accessGroupID db.IDType, queryAccessGroup bool,
) (int64, []cgi.Card, error) {
	// 查询卡记录
	total, cardRecords, err := dac.GetRW().GetCards(
		ctx, mozuID, cards, query,
		cardType, queryCardType, cardFlag, queryCardFlag,
		accessGroupID, queryAccessGroup, offset, limit,
	)
	if err != nil {
		return 0, nil, err
	}

	tempCards := make([]string, 0, len(cards))
	staffIDs := make([]db.IDType, 0, len(cardRecords))
	for i := range cardRecords {
		c := &cardRecords[i]

		sID := c.StaffID
		if sID != db.DefaultStaffID {
			staffIDs = append(staffIDs, sID)
		}

		tempCards = append(tempCards, c.CardNo)
	}
	cards = tempCards

	staffs, err := dac.GetRW().GetStaffsByID(ctx, staffIDs)
	if err != nil {
		return 0, nil, err
	}

	accessGroupRelations, err := dac.GetRW().GetCardAccessGroupRelationByCards(ctx, cards, mozuID)
	if err != nil {
		return 0, nil, err
	}
	cardGroupMap := make(map[string][]db.IDType, len(accessGroupRelations))
	groupMap := make(map[db.IDType]struct{}, len(accessGroupRelations))
	for i := range accessGroupRelations {
		group := accessGroupRelations[i].AccessGroupID
		card := accessGroupRelations[i].CardNo

		groupMap[group] = struct{}{}
		cardGroupMap[card] = append(cardGroupMap[card], group)
	}
	groups := make([]db.IDType, 0, len(groupMap))
	for id := range groupMap {
		groups = append(groups, id)
	}

	groupNames, err := dac.GetRW().GetAccessGroupsBaseInfo(ctx, groups)
	if err != nil {
		return 0, nil, err
	}

	groupDoors, err := dac.GetRW().GetAccessGroupDoors(ctx, groups)
	if err != nil {
		return 0, nil, err
	}

	results := make([]cgi.Card, 0, len(cardRecords))
	for i := range cardRecords {
		c := &cardRecords[i]

		var card cgi.Card
		card.Card = *c

		staff, ok := staffs[c.StaffID]
		if !ok {
			card.Staff.Enable = cgi.StaffUnassigned
		} else {
			card.Staff.Enable = cgi.StaffAssigned
			card.Staff.Staff = staff
			utils.ProcessPersonalInformation(&card.Staff.Staff)
		}

		cardGroups, ok := cardGroupMap[c.CardNo]
		if !ok {
			card.AccessGroups = make([]cgi.AccessGroup, 0)
		} else {
			card.AccessGroups = make([]cgi.AccessGroup, 0, len(cardGroups))
			for i := range cardGroups {
				var cgiGroup cgi.AccessGroup

				cgiGroup.ID = cardGroups[i]
				cgiGroup.Name = groupNames[cgiGroup.ID].Name
				cgiGroup.Doors = utils.GetDoorsBaseInfo(groupDoors[cardGroups[i]])

				card.AccessGroups = append(card.AccessGroups, cgiGroup)
			}
		}

		results = append(results, card)
	}
	return total, results, nil
}
