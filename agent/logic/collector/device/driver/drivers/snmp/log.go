package snmp

import (
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"github.com/gosnmp/gosnmp"
)

const (
	refreshTime = 5 * time.Minute
)

var (
	entryManager *EntryManager
)

func init() {
	entryManager = &EntryManager{
		m: make(IPEntry),
	}

	go entryManager.refresh()
}

// IPKey ip地址和端口
type IPKey struct {
	IP   string
	Port uint16
}

// OIDKey oid
type OIDKey string

// OIDValue oid值
type OIDValue struct {
	Error gosnmp.SNMPError
}

// OIDEntry oid列表
type OIDEntry map[OIDKey]*OIDValue

// IPValue ip值
type IPValue struct {
	sync.RWMutex
	m     OIDEntry
	ipKey IPKey
}

// IPEntry ip列表
type IPEntry map[IPKey]*IPValue

// EntryManager 日志管理器
type EntryManager struct {
	sync.RWMutex
	m IPEntry
}

// GetEntryManager 获取日志管理器
func GetEntryManager() *EntryManager {
	return entryManager
}

func (e *EntryManager) refresh() {
	if e == nil {
		return
	}

	for {
		time.Sleep(refreshTime)
		e.Lock()
		e.m = make(IPEntry)
		e.Unlock()
	}
}

// Log 日志
func (e *EntryManager) Log(ip string, port uint16, oid string, error gosnmp.SNMPError) {
	if e == nil {
		return
	}

	e.Lock()
	defer e.Unlock()

	ipKey := IPKey{
		IP:   ip,
		Port: port,
	}
	ipValue, ok := e.m[ipKey]
	if !ok {
		ipValue = &IPValue{
			m:     make(OIDEntry),
			ipKey: ipKey,
		}
		e.m[ipKey] = ipValue
	}

	ipValue.Log(oid, error)
}

// Log 日志
func (i *IPValue) Log(oid string, error gosnmp.SNMPError) {
	if i == nil {
		return
	}

	i.Lock()
	v, ok := i.m[OIDKey(oid)]
	if !ok {
		v = &OIDValue{Error: error}
		i.m[OIDKey(oid)] = v
	} else {
		if v.Error == error {
			i.Unlock()
			return
		} else {
			v.Error = error
		}
	}
	ip := i.ipKey.IP
	port := i.ipKey.Port
	i.Unlock()

	log.Warnf(
		"ip: %+v, port: %+v, oid: %+v, snmp response error status: %+v",
		ip, port, oid, error,
	)
}
