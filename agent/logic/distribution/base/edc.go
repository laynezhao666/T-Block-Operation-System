package base

import (
	"agent/entity/consts"
	"agent/entity/definition"
)

var (
	CpuLoad      = "CpuLoad"
	DisUsePer    = "DisUsePer"
	DiskAvail    = "DiskAvail"
	DiskTotal    = "DiskTotal"
	MemAvail     = "MemAvail"
	MemTotal     = "MemTotal"
	MemUsePer    = "MemUsePer"
	NetDownFlow  = "NetDownFlow"
	NetDownRate  = "NetDownRate"
	NetUpFlow    = "NetUpFlow"
	NetUpRate    = "NetUpRate"
	PowerFault_1 = "PowerFault_1"
	PowerFault_2 = "PowerFault_2"
)
// EdcMonitorPoints 获取edc监控点
func EdcMonitorPoints(edcGid string) definition.DataPointIDsType {
	return definition.DataPointIDsType{
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + CpuLoad),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + DisUsePer),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + DiskAvail),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + DiskTotal),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + MemAvail),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + MemTotal),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + MemUsePer),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + NetDownFlow),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + NetDownRate),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + NetUpFlow),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + NetUpRate),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + PowerFault_1),
		definition.DataPointIDType(edcGid + consts.DefaultIDSep + PowerFault_2),
	}
}
