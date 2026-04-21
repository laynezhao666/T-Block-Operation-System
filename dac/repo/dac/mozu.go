package dac

import (
	"context"

	"dac/entity/model/db"

	"gorm.io/gorm"
)

// GetMozuWithSameBuildings 获取与指定模组同楼栋的所有模组。
func (d *impl) GetMozuWithSameBuildings(ctx context.Context, mozuID int) ([]db.Mozu, error) {
	var mozus []db.Mozu
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m db.Mozu
		e := tx.Scopes(withMozuID(mozuID)).First(&m).Error
		if e != nil {
			return e
		}

		return tx.Scopes(withBuildingMID(m.BuildingMID)).Find(&mozus).Error
	})
	return mozus, err
}
