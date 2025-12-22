package rpc

import (
	"sync"

	"etrpc-go/log"
	"etrpc-go/util/httputil"
	cmdbPb "trpcprotocol/cmdb"

	"trpc.group/trpc-go/trpc-go"
)

var (
	cmdbOnce sync.Once
	cmdbSvc  *CmdbSvc
)

// CmdbSvc 测点数据查询服务
type CmdbSvc struct {
	cmdbCli cmdbPb.ConfigQueryClientProxy
}

// GetCmdbSvc GetCmdbSvc
func GetCmdbSvc() *CmdbSvc {
	cmdbOnce.Do(func() {
		cmdbSvc = &CmdbSvc{
			cmdbCli: cmdbPb.NewConfigQueryClientProxy(),
		}
	})
	return cmdbSvc
}

// GetMozuInfoList 获取mozu信息列表
func (s *CmdbSvc) GetMozuInfoList() ([]*cmdbPb.RspGetMozuInfo_MozuInfo, error) {
	req := &cmdbPb.ReqGetMozuInfo{
		MozuId: []int32{},
	}
	rsp, err := s.cmdbCli.GetMozuInfo(trpc.BackgroundContext(), req, httputil.GetPbCallOption())
	if err != nil {
		log.Errorf("GetMozuInfoList failed, err: %v", err)
		return nil, err
	}
	return rsp.List, nil
}

// GetDeviceEntity 获取设备实体
func (s *CmdbSvc) GetDeviceEntity(mozuId int64, page, size int) (*cmdbPb.RspGetDeviceEntity, error) {
	req := &cmdbPb.ReqGetDeviceEntity{
		MozuId: []int32{int32(mozuId)},
		Page:   int32(page),
		Size:   int32(size),
	}
	rsp, err := s.cmdbCli.GetDeviceEntity(trpc.BackgroundContext(), req)
	if err != nil {
		log.Errorf("GetDeviceEntity failed, mozuId:%d, err: %v", mozuId, err)
		return nil, err
	}
	return rsp, nil
}
