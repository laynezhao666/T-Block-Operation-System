package cache

import (
	"context"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/repo/dac"
)

// refreshCardStaffTime 卡-员工映射缓存刷新间隔
const (
	refreshCardStaffTime = time.Minute
)

// DeleteStaff 从缓存中删除指定员工关联的所有卡映射。
func (c *Cache) DeleteStaff(mozuID string, staffID db.IDType) {
	c.mozuCardStaffMutex.Lock()
	defer c.mozuCardStaffMutex.Unlock()

	cardStaffMap, ok := c.mozuCardStaffMap[mozuID]
	if !ok {
		return
	}

	deleteCards := make([]string, 0)
	for card, s := range cardStaffMap {
		if s.ID != staffID {
			continue
		}
		deleteCards = append(deleteCards, card)
	}

	for _, card := range deleteCards {
		delete(cardStaffMap, card)
	}
}

// UpdateStaff 更新缓存中指定员工的信息。
// 遍历该员工关联的所有卡，更新对应的员工数据。
func (c *Cache) UpdateStaff(staff db.Staff, mozuID string) {
	c.mozuCardStaffMutex.Lock()
	defer c.mozuCardStaffMutex.Unlock()

	cardStaffMap, ok := c.mozuCardStaffMap[mozuID]
	if !ok {
		return
	}

	updateCards := make([]string, 0)
	for card, s := range cardStaffMap {
		if s.ID != staff.ID {
			continue
		}
		updateCards = append(updateCards, card)
	}

	for _, card := range updateCards {
		cardStaffMap[card] = staff
	}
}

// GetCardStaffMap 获取指定墨组下的卡号到员工信息的映射副本。
// 返回的是深拷贝，修改不会影响缓存。
func (c *Cache) GetCardStaffMap(mozuID string) map[string]db.Staff {
	c.mozuCardStaffMutex.RLock()
	defer c.mozuCardStaffMutex.RUnlock()

	cardStaffMap, ok := c.mozuCardStaffMap[mozuID]
	if !ok {
		return nil
	}

	m := make(map[string]db.Staff, len(cardStaffMap))
	for k, v := range cardStaffMap {
		m[k] = v
	}
	return m
}

// refreshCardStaffMap 从数据库刷新卡-员工映射缓存。
func (c *Cache) refreshCardStaffMap(ctx context.Context) {
	_, m, err := dac.GetRW().GetAllCardStaffMap(ctx)
	if err != nil {
		config.Log.Warnf("get card staff map error: %v", err)
	}

	c.mozuCardStaffMutex.Lock()
	defer c.mozuCardStaffMutex.Unlock()

	c.mozuCardStaffMap = m
}

// refreshCardStaffMapLoop 定时循环刷新卡-员工映射缓存。
func (c *Cache) refreshCardStaffMapLoop(ctx context.Context) {
	for {
		c.refreshCardStaffMap(ctx)
		select {
		case <-time.After(refreshCardStaffTime):
			break
		case <-ctx.Done():
			config.Log.Info("stop refresh card staff loop.")
			return
		}
	}
}
