// Package group 提供门组的增删改查业务逻辑。
package group

import (
	"context"

	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/repo/dac"
)

// GetGroupDoors 获取门组下的所有门信息
func GetGroupDoors(ctx context.Context, groupID int) ([]db.Door, error) {
	doorRecords, err := dac.GetRW().GetGroupDoors(ctx, groupID)
	if err != nil {
		return nil, err
	}

	for i := range doorRecords {
		utils.ProcessDBDoor(&doorRecords[i])
	}

	return doorRecords, nil
}
