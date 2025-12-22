package osal

import (
	"log"
	"sync"
	"time"
)

const (
	PacketLogKey   = "TBMON_PACKET_LOG"
	kAutoResetTime = 1 * time.Minute
)

// ResetArgs 重置参数
type ResetArgs struct {
	IDs []string
}

// EnvVar 环境变量
type EnvVar struct {
	mu           sync.Mutex
	varMap       map[string]int
	varsMap      map[string]map[string]int
	keepaliveMap map[string]VarKeepalive
}

// VarKeepalive 变量存活
type VarKeepalive struct {
	Keepalive     int
	LastHeartbeat time.Time
}

var instance *EnvVar

// Instance 返回环境变量实例
func Instance() *EnvVar {
	if instance == nil {
		instance = &EnvVar{
			varMap:       make(map[string]int),
			varsMap:      make(map[string]map[string]int),
			keepaliveMap: make(map[string]VarKeepalive),
		}
	}
	return instance
}

// AutoResetLogPacket 自动重置日志
func (e *EnvVar) AutoResetLogPacket(args *ResetArgs) {
	time.Sleep(kAutoResetTime)

	e.mu.Lock()
	defer e.mu.Unlock()

	if vars, ok := e.varsMap[PacketLogKey]; ok {
		for _, id := range args.IDs {
			log.Printf("disable channel \"%s\" log", id)
			delete(vars, id)
		}
	}
}

// Get 获取环境变量
func (e *EnvVar) Get(name string) (int, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	val, ok := e.varMap[name]
	return val, ok
}

// GetWithID 获取环境变量
func (e *EnvVar) GetWithID(name, id string) (int, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if vars, ok := e.varsMap[name]; ok {
		val, ok := vars[id]
		return val, ok
	}
	return 0, false
}

// Set 设置环境变量
func (e *EnvVar) Set(name string, val int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.varMap[name] = val
}

// BatchSet 批量设置环境变量
func (e *EnvVar) BatchSet(name string, ids []string, val int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if val != 0 && name == PacketLogKey {
		go e.AutoResetLogPacket(&ResetArgs{IDs: ids})
	}

	if _, ok := e.varsMap[name]; !ok {
		e.varsMap[name] = make(map[string]int)
	}
	for _, id := range ids {
		log.Printf("enable channel: \"%s\" log: %d", id, val)
		e.varsMap[name][id] = val
	}
}

// Delete 删除环境变量
func (e *EnvVar) Delete(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.varMap, name)
	delete(e.varsMap, name)
}

// SetKeepAlive 设置环境变量心跳
func (e *EnvVar) SetKeepAlive(name string, keepalive int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.keepaliveMap[name] = VarKeepalive{
		Keepalive:     keepalive,
		LastHeartbeat: time.Now(),
	}
}

// CheckKeepalive 检查环境变量心跳
func (e *EnvVar) CheckKeepalive() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	for name, varKeepalive := range e.keepaliveMap {
		if now.Sub(varKeepalive.LastHeartbeat) >= time.Duration(varKeepalive.Keepalive)*time.Second {
			e.Delete(name)
			delete(e.keepaliveMap, name)
		}
	}
}
