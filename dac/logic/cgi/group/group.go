// Package group 提供门组的增删改查业务逻辑。
package group

import (
	"context"

	"dac/entity/model/db"
	"dac/repo/dac"
)

// Create 创建门组并关联指定的门
func Create(ctx context.Context, mozuID string, groupName string, doorIDs []int) error {
	group := db.DoorGroup{
		Name:   groupName,
		MozuID: mozuID,
	}
	return dac.GetRW().AddGroup(ctx, group, doorIDs)
}

// Update 更新门组名称和关联的门
func Update(ctx context.Context, groupID db.IDType, groupName string, doorIDs []db.IDType) error {
	group := db.DoorGroup{ID: groupID, Name: groupName}
	return dac.GetRW().UpdateGroup(ctx, group, doorIDs)
}

// Delete 删除门组
func Delete(ctx context.Context, id db.IDType) error {
	return dac.GetRW().DeleteGroup(ctx, id)
}
