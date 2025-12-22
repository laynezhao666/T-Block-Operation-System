package change

import (
	"sync"
)

type pointCount struct {
	currentNotChangedPointsNum uint64
	currentCallbackPointsNum   uint64
	sync.RWMutex
}
// Add 增加未回调和已回调的数量
func (p *pointCount) Add(notChanged, callback uint64) {
	p.Lock()
	defer p.Unlock()

	p.currentNotChangedPointsNum += notChanged
	p.currentCallbackPointsNum += callback
}
// Get 获取未回调和已回调的数量
func (p *pointCount) Get() (uint64, uint64) {
	p.Lock()
	defer p.Unlock()

	n, m := p.currentNotChangedPointsNum, p.currentCallbackPointsNum
	p.currentNotChangedPointsNum, p.currentCallbackPointsNum = 0, 0
	return n, m
}
