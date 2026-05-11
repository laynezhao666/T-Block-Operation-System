package cgi

import (
	"agent/entity/config"
	"agent/entity/errcode"
	"agent/logic/cm"
	pb "trpcprotocol/agent"
)

// ProfileHandle tbox信息接口
func ProfileHandle() (*pb.ProfileRsp, error) {
	var gid, code string
	if config.GetRB().IsGatewayMode() {
		gid = ""
	}
	tboxGids := cm.Worker().GetTboxDeviceGids()
	if len(tboxGids) > 0 {
		gid = string(tboxGids[0])
	}
	codes := config.GetRB().Task.Local.Devs
	if len(codes) > 0 {
		code = codes[0]
	}
	data := &pb.Profile{
		Gid:  gid,
		Code: code,
	}
	return &pb.ProfileRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    data,
	}, nil
}
