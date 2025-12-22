package common

import (
	"cgi/repo/db"
	"common/entity/model"
	"context"
	"etrpc-go/util/httputil"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/http"

	pb "trpcprotocol/cgi"
)

// ICommonApi 通用接口定义
type ICommonApi interface {
	// ExportData 导出数据
	ExportData(ctx context.Context, req *pb.ReqCommonExportData) (*emptypb.Empty, error)
	// GetKeyDict 获取Key字典列表
	GetKeyDict(ctx context.Context, req *pb.ReqCommonGetDict) (*pb.RspCommonGetKeyDict, error)
	// GetKvDict 获取KV字典类别
	GetKvDict(ctx context.Context, req *pb.ReqCommonGetDict) (*pb.RspCommonGetKvDict, error)
}

// NewCommonApi 创建一个统一API
func NewCommonApi() ICommonApi {
	commonDao := db.NewCommonDao()
	deviceEntityDao := db.NewDeviceEntityDao()
	return &commonApi{
		exportFuncMap: map[string]string{
			"device_entity":  "Cmdb/GetDeviceEntity",
			"device_point":   "Cmdb/GetDevicePoint",
			"point_data":     "Data/Query",
			"alarm_list":     "alarm/server/GetAlarmList",
			"alarm_strategy": "alarm/server/GetStrategy",
			"alarm_validate": "alarm/server/GetValidate",
		},
		keyDicFuncMap: map[string]db.GetKeyDicFunc{
			"application_type_en_list": commonDao.TableKeyDicFunc("t_device_application_type",
				"application_type_en", false),
			"application_type_zh_list": commonDao.TableKeyDicFunc("t_device_application_type",
				"application_type_zh", false),
			"device_type_en_list": commonDao.TableKeyDicFunc("t_device_type",
				"device_type_en", false),
			"device_type_zh_list": commonDao.TableKeyDicFunc("t_device_type",
				"device_type_zh", false),
			"device_number_list": deviceEntityDao.KeyDicFunc(
				func(item *model.DeviceEntity) string { return item.DeviceNumber }),
		},
		kvDicFuncMap: map[string]db.GetKvDicFunc{
			"application_type_kv": commonDao.TableKvDicFunc(
				"t_device_application_type", "application_type_en", "application_type_zh", false),
			"device_type_kv": commonDao.TableKvDicFunc(
				"t_device_type", "device_type_en", "device_type_zh", false),
			"device_number_kv": deviceEntityDao.KvDicFunc(func(item *model.DeviceEntity) string { return item.DeviceNumber },
				func(item *model.DeviceEntity) string { return item.DeviceGid }),
			"device_gid_kv": deviceEntityDao.KvDicFunc(func(item *model.DeviceEntity) string { return item.DeviceGid },
				func(item *model.DeviceEntity) string { return item.DeviceNumber }),
			"mozu_application_type_kv": deviceEntityDao.KvDicFunc(func(item *model.DeviceEntity) string { return item.ApplicationTypeEn },
				func(item *model.DeviceEntity) string { return item.ApplicationTypeZh }),
		},
		commonDao:       commonDao,
		deviceEntityDao: deviceEntityDao,
	}
}

type commonApi struct {
	exportFuncMap map[string]string
	keyDicFuncMap map[string]db.GetKeyDicFunc
	kvDicFuncMap  map[string]db.GetKvDicFunc

	commonDao       db.ICommonDao
	deviceEntityDao db.IDeviceEntityDao
}

type exportRsp struct {
	List []map[string]any `json:"list"`
}

func (c commonApi) ExportData(ctx context.Context, req *pb.ReqCommonExportData) (*emptypb.Empty, error) {
	if exportPath, ok := c.exportFuncMap[req.ExportType]; ok {
		// 调用接口获取数据
		rsp := &httputil.ResponseEntity[exportRsp]{}
		err := httputil.PostJson(ctx, fmt.Sprintf("http://idc-tbos-cgi/%s", exportPath), nil, req.Param, rsp)
		if err != nil {
			return nil, errors.Wrapf(err, "fecth data fail")
		}
		if rsp.Code != 0 {
			return nil, fmt.Errorf("fecth data fail, rsp: %v", rsp)
		}
		// 生成excel
		excel := excelize.NewFile()
		for j, field := range req.Fields {
			// 逐列来写
			pos := fmt.Sprintf("%c%d", 'A'+j, 1)
			err := excel.SetCellValue("Sheet1", pos, field.FieldCn)
			if err != nil {
				return nil, errors.Wrapf(err, "write excel fail at pos:[%s], val:[%s]", pos, field.FieldCn)
			}
			for i, item := range rsp.Data.List {
				pos := fmt.Sprintf("%c%d", 'A'+j, i+2)
				err := excel.SetCellValue("Sheet1", pos, item[field.FieldEn])
				if err != nil {
					return nil, errors.Wrapf(err, "write excel fail at pos:[%s], val:[%v]", pos, item[field.FieldEn])
				}
			}
		}
		head := http.Head(ctx)
		msg := trpc.Message(ctx)
		msg.WithSerializationType(codec.SerializationTypeUnsupported)
		head.Response.Header().Add("Content-Type", "application/octet-stream;Charset=utf-8")
		head.Response.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s",
			url.QueryEscape(req.FileName)))
		head.Response.Header().Add("Content-Transfer-Encoding", "binary")
		if err := excel.Write(head.Response); err != nil {
			return nil, errors.Wrapf(err, "convert excel to data stream fail")
		}
		return nil, err
	} else {
		return nil, fmt.Errorf("unsupport export_type [%s]", req.ExportType)
	}
}

func (c commonApi) GetKeyDict(ctx context.Context, req *pb.ReqCommonGetDict) (
	*pb.RspCommonGetKeyDict, error) {
	if dicFunc, ok := c.keyDicFuncMap[req.DicType]; ok {
		// 查询数据
		dicList, err := dicFunc(ctx, req.Filter, req.MozuId)
		if err != nil {
			return nil, errors.Wrapf(err, "request data fail")
		}
		return &pb.RspCommonGetKeyDict{
			List: dicList,
		}, nil
	} else {
		return nil, fmt.Errorf("unsupport dic_type [%s]", req.DicType)
	}
}

func (c commonApi) GetKvDict(ctx context.Context, req *pb.ReqCommonGetDict) (
	*pb.RspCommonGetKvDict, error) {
	if dicFunc, ok := c.kvDicFuncMap[req.DicType]; ok {
		// 查询数据
		kvs, err := dicFunc(ctx, req.Filter, req.MozuId)
		if err != nil {
			return nil, errors.Wrapf(err, "request data fail")
		}
		return &pb.RspCommonGetKvDict{
			Kvs: kvs,
		}, nil
	} else {
		return nil, fmt.Errorf("unsupport dic_type [%s]", req.DicType)
	}
}
