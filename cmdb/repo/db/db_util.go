package db

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"trpc.group/trpc-go/trpc-go/log"
)

// TransactionUpdate 事务更新数据
func TransactionUpdate[T any, R int | int32 | int64 | uint | uint32 | uint64](session *gorm.DB, addList []*T, delIdList []R, desc string) error {
	if len(addList) > 0 || len(delIdList) > 0 {
		// 事务更新
		return session.Transaction(func(tx *gorm.DB) error {
			if len(delIdList) > 0 {
				// 如果单次删除的数据量太大,分批进行删除
				for _, chunk := range lo.Chunk(delIdList, 5000) {
					if err := tx.Delete(new(T), chunk).Error; err != nil {
						return errors.Wrapf(err, "delete old %s failed", desc)
					}
				}
				log.InfoContextf(session.Statement.Context, "delete old [%s] success, total [%d]", desc, len(delIdList))
			}
			if len(addList) > 0 {
				// 如果单次插入的数据量太大,分批进行插入
				if err := tx.CreateInBatches(addList, 1000).Error; err != nil {
					return errors.Wrapf(err, "save new %s failed", desc)
				}
				log.InfoContextf(session.Statement.Context, "save new [%s] success, total [%d]", desc, len(addList))
			}
			return nil
		})
	}
	return nil
}
