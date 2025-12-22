// Package cache 监听模组数据是否出现变化
package cache

import (
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"
	"etrpc-go/log"
	"fmt"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go"
)

var mozuVerMap sync.Map

// InitCache 初始化缓存
func InitCache(ctx context.Context) {
	loadMozuVer(true)
	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-ctx.Done():
				log.Infof("refresh cache stoped")
				return
			case <-ticker.C:
				loadMozuVer(false)
				ticker.Reset(time.Second * 30)
			}
		}
	}()
}

func loadMozuVer(firstLoad bool) {
	db := tgorm.GetDB(consts.TbosMysqlName)
	res := make([]*model.MozuInfo, 0)
	if err := db.Select("mozu_id, publish_version").Find(&res).Error; err != nil {
		if firstLoad {
			panic(fmt.Sprintf("cache get mozu ver info fail, err: %v", err))
		} else {
			log.AlarmContextf(trpc.BackgroundContext(), "cache get mozu ver info fail, err: %v", err)
		}
		return
	}
	for _, newMozuInfo := range res {
		if oldVer, ok := mozuVerMap.Load(newMozuInfo.MozuId); !ok || newMozuInfo.PublishVersion != oldVer {
			mozuVerMap.Store(newMozuInfo.MozuId, newMozuInfo.PublishVersion)
		}
	}
}
