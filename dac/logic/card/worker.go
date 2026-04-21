// Package card 提供门禁卡的定时清理功能，自动禁用过期临时卡。
package card

import (
	"context"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/repo/dac"

	"dac/entity/utils/ttime"
	"gorm.io/gorm"
)

// cleanTime 过期卡清理任务执行间隔
const (
	cleanTime = time.Minute * 5
)

// w 全局Worker单例
var (
	w = new(Worker)
)

// Worker 过期卡清理工作器
type Worker struct {
}

// getWorker 获取全局Worker实例
func getWorker() *Worker {
	return w
}

// Init 初始化卡片清理模块并启动后台清理协程
func Init(ctx context.Context) {
	getWorker().start(ctx)
}

// start 启动后台清理协程
func (w *Worker) start(ctx context.Context) {
	go w.cleanExpiredCardsLoop(ctx)
}

// cleanExpiredCards 扫描并禁用所有已过期的临时卡
func (w *Worker) cleanExpiredCards(ctx context.Context) {
	t := ttime.GetNowUTC().Unix()
	disableCardsMap := make(map[string][]string)
	_, err := dac.GetRW().GetAllCards(ctx, func(tx *gorm.DB, cards []db.Card) ([]db.Card, error) {
		for i := range cards {
			c := &cards[i]
			if c.ValidTime > 0 && c.CardType == db.CardTypeTemporary &&
				c.CardFlag == db.CardFlagEnable && t >= c.ValidTime {
				disableCardsMap[c.MozuID] = append(disableCardsMap[c.MozuID], c.CardNo)
			}
		}

		var err error
		for mozuID, disableCards := range disableCardsMap {
			if err = dac.UpdateCardsFlag(tx, disableCards, mozuID, db.CardFlagDisable, func(tx *gorm.DB) error {
				return updateFlagInController(tx, disableCards, mozuID, db.CardFlagDisable)
			}); err != nil {
				break
			}
		}
		return cards, err
	})
	if err != nil {
		config.Log.Warnf("disable expired cards %+v error: %v", disableCardsMap, err)
		return
	}

	if len(disableCardsMap) == 0 {
		return
	}

	config.Log.Infof("disable expired cards %+v success.", disableCardsMap)
}

// cleanExpiredCardsLoop 后台循环执行过期卡清理任务
func (w *Worker) cleanExpiredCardsLoop(ctx context.Context) {
	for {
		w.cleanExpiredCards(ctx)
		select {
		case <-time.After(cleanTime):
			break
		case <-ctx.Done():
			config.Log.Infof("stop clean expired cards loop.")
			return
		}
	}
}
