package dac

import (
	"context"

	"dac/entity/model/rt"

	"gorm.io/gorm"
)

// UpdateControllerAndDoorGIDsByCode 根据采集编码批量更新控制器和门的GID（impl 方法）。
func (d *impl) UpdateControllerAndDoorGIDsByCode(ctx context.Context, codeGIDs rt.CodeGIDMapType) error {
	if len(codeGIDs) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return UpdateControllerAndDoorGIDsByCode(tx, codeGIDs)
	})
}

// UpdateControllerAndDoorGIDsByCode 根据采集编码批量更新控制器和门的GID。
func UpdateControllerAndDoorGIDsByCode(tx *gorm.DB, codeGIDs rt.CodeGIDMapType) error {
	if len(codeGIDs) == 0 {
		return nil
	}
	var err error

	if err = updateControllerGIDsByCode(tx, codeGIDs); err != nil {
		return err
	}

	if err = updateDoorGIDsByCode(tx, codeGIDs); err != nil {
		return err
	}

	return nil
}
