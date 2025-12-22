package utils

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/cm"
	"strings"
	pb "trpcprotocol/agent"
)

// GetMappingObject 获取映射完整对象
func GetMappingObject(mapping string) (map[string]*pb.RefObject, error) {
	mpList := strings.Split(mapping, consts.MpListSplitChar)

	param := make(map[string]*pb.RefObject)
	for _, v := range mpList {
		parts := strings.Split(v, consts.MpExpressionSplitChar)
		if len(parts) != 2 {
			continue
			//return nil, errors.New("mapping format err")
		}
		// 补充1297036692686905345.EaP的信息
		refParts := strings.Split(parts[1], consts.MpExpressionRefSplitChar)
		if len(refParts) != 2 {
			continue
			//return nil, errors.New("mapping format err")
		}
		objectGid := refParts[0]
		device, ok := cm.Worker().GetDeviceByGid(definition.DeviceGidType(objectGid))
		if !ok {
			continue
			//return nil, errors.New("mapping format err")
		}
		ro := &pb.RefObject{
			// 后续可能引用标准点
			Type:        "collect",
			Id:          device.ID,
			Gid:         objectGid,
			PointNameEn: refParts[1],
		}
		param[parts[0]] = ro
	}
	return param, nil
}
