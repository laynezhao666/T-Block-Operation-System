package service

import (
	"agent/logic/collector/dispatcher"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/errcode"
	"agent/entity/model"
	"agent/logic/cgi/usr"
	"agent/logic/cm"
	utils2 "agent/logic/cm/utils"
	"agent/logic/collector/rtdb"
	cmodel "agent/logic/collector/rtdb/model"
	"agent/logic/logfile"
	"agent/logic/network"
	"agent/logic/std"
	"agent/utils"

	pb "trpcprotocol/agent"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	filePathSep string = "/"
)

// ConfigManager 配置管理
type ConfigManager struct{}

func (c ConfigManager) TIotControl(ctx context.Context, req *pb.TIotControlReq) (*pb.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (c ConfigManager) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.RspGetTemplate, error) {
	//TODO implement me
	panic("implement me")
}

func (c ConfigManager) GetTemplateFromCloud(ctx context.Context, req *pb.GetTemplateRequest) (*pb.RspGetTemplate, error) {
	//TODO implement me
	panic("implement me")
}

/***************** devices group *****************/

// ProcessAllDevices /tbcm/devices GET获取所有设备 DELETE根据一组id删除多个设备
func (c ConfigManager) ProcessAllDevices(ctx context.Context, req *pb.ProcessAllDevicesReq) (*pb.ProcessAllDevicesRsp, error) {
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	rsp := &pb.ProcessAllDevicesRsp{
		RspGet: &pb.ProcessAllDevicesRsp_RspGet{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
		},
		RspDel: &pb.ProcessAllDevicesRsp_RspDel{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
		},
	}
	switch head.Request.Method {
	case "GET":
		devices := cm.Worker().GetAllDevices()
		for _, d := range devices {
			// 暂时办法，解决通道串口通道显示为"/usr/dev/serial/com7"而非"COM7"的问题
			l := strings.Split(d.ChData.ChannelID, "/")
			convertedChannelId := strings.ToUpper(l[len(l)-1])
			dev := pb.CollectDevice{
				Gid:  string(d.Gid),
				Id:   string(d.ID),
				Name: d.Name,
				Type: d.TypeEn,
				Tpl: &pb.Template{
					Tplnm:   d.TemplateData.TemplateName,
					Tplpath: d.TemplateData.TemplatePath,
				},
				Channel: &pb.Channel{
					Addr: d.ChData.Address,
					// Chid: v.ChData.ChannelID,
					Chid:         convertedChannelId,
					Chparams:     d.ChData.ChannelParams,
					Chtype:       d.ChData.Chtype,
					Timeout:      fmt.Sprintf("%v", d.ChData.TimeoutMs),
					CmdIntetrval: fmt.Sprintf("%v", d.ChData.CmdInterval),
					WaitTime:     fmt.Sprintf("%v", d.ChData.WaitTimeMs),
				},
				MozuId: int32(d.MozuID),
			}
			rsp.RspGet.Data = append(rsp.RspGet.Data, &dev)
			sort.Slice(rsp.RspGet.Data, func(i, j int) bool {
				return rsp.RspGet.Data[i].Id < rsp.RspGet.Data[j].Id
			})
		}
	case "DELETE":
		ids := req.GetReqDel().GetIds()
		cm.Worker().DeleteDevicesByIds(ids...)
		err := cm.Worker().SaveCurrentDevicesConfig(definition.CollectorDeviceTypeTBox)
		if err != nil {
			return nil, errs.New(errcode.ErrSaveConfigFail, fmt.Sprintf("save devices config fail: %v", err))
		}
		cm.NotifyConfigChange()
		rsp.RspDel.Data = ids
	default:
		return nil, errs.New(
			errcode.ErrCgiHttpMethodNotSupported,
			fmt.Sprintf("method <%v> not support", head.Request.Method),
		)
	}
	return rsp, nil
}

// GetCollectDeviceVersion /tbcm/devices/version GET获取配置版本
func (c ConfigManager) GetCollectDeviceVersion(ctx context.Context, req *emptypb.Empty) (*pb.GetCollectDeviceVersionRsp, error) {
	// TODO 待获取配置版本
	return &pb.GetCollectDeviceVersionRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data: &pb.GetCollectDeviceVersionRsp_VersionData{
			Version: "1.0",
		},
	}, nil
}

// BatchGetCollectDevices /tbcm/devices/get POST根据一组gid批量获取设备配置
func (c ConfigManager) BatchGetCollectDevices(ctx context.Context, req *pb.BatchGetReq) (*pb.GetCollectDevicesRsp, error) {
	rsp := &pb.GetCollectDevicesRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	ids := req.GetIds()
	devices, ok, notFoundIds := cm.Worker().GetDevicesByIds(ids)
	if !ok {
		return rsp, errs.New(errcode.ErrCgiParamInvalid, fmt.Sprintf("some devices not found: %v", notFoundIds))
	}
	rsp.Data = make([]*pb.CollectDevice, 0, len(devices))
	for _, d := range devices {
		rsp.Data = append(rsp.Data, &pb.CollectDevice{
			Gid:  string(d.Gid),
			Id:   string(d.ID),
			Name: d.Name,
			Type: d.TypeEn,
			Tpl: &pb.Template{
				Tplnm:   d.TemplateData.TemplateName,
				Tplpath: d.TemplateData.TemplatePath,
			},
			Channel: &pb.Channel{
				Addr:         d.ChData.Address,
				Chid:         d.ChData.ChannelID,
				Chparams:     d.ChData.ChannelParams,
				Chtype:       d.ChData.Chtype,
				Timeout:      fmt.Sprintf("%v", d.ChData.TimeoutMs),
				CmdIntetrval: fmt.Sprintf("%v", d.ChData.CmdInterval),
				WaitTime:     fmt.Sprintf("%v", d.ChData.WaitTimeMs),
			},
		})
	}
	sort.Slice(rsp.Data, func(i, j int) bool {
		return rsp.Data[i].Id < rsp.Data[j].Id
	})
	return rsp, nil
}

// BatchCreateDevices /tbcm/devices/batch POST批量创建设备
func (c ConfigManager) BatchCreateDevices(ctx context.Context, req *pb.BatchDevicesReq) (*pb.BatchDevicesRsp, error) {
	return c.BatchUpdateDevices(ctx, req)
}

// BatchUpdateDevices /tbcm/devices/batch-update POST批量更新设备
func (c ConfigManager) BatchUpdateDevices(ctx context.Context, req *pb.BatchDevicesReq) (*pb.BatchDevicesRsp, error) {
	if config.GetRB().Project.Source != "local" {
		return &pb.BatchDevicesRsp{}, errs.New(errcode.ErrBadRequest, "当前不是local模式，无法执行该操作")
	}
	worker := cm.Worker()
	rsp := &pb.BatchDevicesRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	var err error
	ids := make([]string, 0, len(req.Devices))
	for _, device := range req.Devices {
		if device == nil {
			continue
		}
		// if device.Gid == "" {
		// 	device.Gid = string(cm.Worker().GetNextDeviceGid())
		// }
		ids = append(ids, device.Id)
		var cmdInterval, waitTime, timeout int
		cmdIntervalStr := device.GetChannel().GetCmdIntetrval()
		if cmdIntervalStr == "" {
			cmdInterval = 0
		} else {
			cmdInterval, err = strconv.Atoi(cmdIntervalStr)
			if err != nil {
				return nil, errs.New(errcode.ErrCgiParamInvalid, "cmd interval invalid")
			}
		}
		waitTimeStr := device.GetChannel().GetWaitTime()
		if waitTimeStr == "" {
			waitTime = 0
		} else {
			waitTime, err = strconv.Atoi(waitTimeStr)
			if err != nil {
				return nil, errs.New(errcode.ErrCgiParamInvalid, "wait time invalid")
			}
		}
		timeoutStr := device.GetChannel().GetTimeout()
		if timeoutStr == "" {
			timeout = 0
		} else {
			timeout, err = strconv.Atoi(timeoutStr)
			if err != nil {
				return nil, errs.New(errcode.ErrCgiParamInvalid, "timeoutStr invalid")
			}
		}
		worker.SetDevice(definition.DeviceGidType(device.Id), model.Device{
			Gid:  worker.GetNextDeviceGid(),
			Name: device.GetName(),
			ID:   device.GetId(),
			ChData: model.ChannelData{
				ChannelID:     device.GetChannel().GetChid(),
				ChannelParams: device.GetChannel().GetChparams(),
				Address:       device.GetChannel().GetAddr(),
				CmdInterval:   cmdInterval,
				WaitTimeMs:    waitTime,
				TimeoutMs:     timeout,
				Chtype:        device.GetChannel().GetChtype(),
			},
			TemplateData: model.TemplateInfo{
				TemplateName: device.GetTpl().GetTplnm(),
				TemplatePath: device.GetTpl().GetTplpath(),
			},
		})
	}
	if len(req.Devices) > 0 {
		err := worker.SaveCurrentDevicesConfig(definition.CollectorDeviceTypeTBox)
		if err != nil {
			return nil, errs.New(errcode.ErrSaveConfigFail, fmt.Sprintf("save devices config fail: %v", err))
		}
		cm.NotifyConfigChange()
	}
	rsp.Data = ids
	return rsp, nil
}

// PostDevicesTemplates /tbcm/devices/tpl POST从模板导入设备信息
func (c ConfigManager) PostDevicesTemplates(ctx context.Context, req *emptypb.Empty) (*pb.PostDevicesTemplatesRsp, error) {
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	request := head.Request
	if err := request.ParseMultipartForm(100 * (1 << 20)); err != nil {
		return nil, errs.New(errcode.ErrCgiTemplateFileFail, "devices file build fail: parse multipartForm failed")
	}
	if request.MultipartForm == nil || request.MultipartForm.File == nil {
		return nil, errs.New(errcode.ErrCgiTemplateFileFail, "devices file build fail: parse multipartForm empty")
	}
	filesHeaders, ok := head.Request.MultipartForm.File["files"]
	if !ok {
		return nil, errs.New(errcode.ErrCgiTemplateFileFail, "devices file build fail: get files failed")
	}
	log.Info("DevicePost, get files ok")

	err := utils2.ProcessImportDevice(filesHeaders)
	if err != nil {
		return nil, errs.New(errcode.ErrCgiTemplateFileFail, fmt.Sprintf("devices file build fail: %v", err.Error()))
	}
	return nil, nil
}

// DownloadDevices /tbcm/devices/download GET下载设备信息
func (c ConfigManager) DownloadDevices(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if config.GetRB().Project.Source != "local" {
		return &emptypb.Empty{}, errs.New(errcode.ErrBadRequest, "当前不是local模式，无法执行该操作")
	}
	return &emptypb.Empty{}, utils2.ExportAllDevices(ctx)
}

// ExportDeviceTemplate /tbcm/tpls/device GET下载设备表格模板
func (c ConfigManager) ExportDeviceTemplate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if config.GetRB().Project.Source != "local" {
		return &emptypb.Empty{}, errs.New(errcode.ErrBadRequest, "当前不是local模式，无法执行该操作")
	}
	path := filepath.Join(consts.ProjectPath, consts.EmptyDevicesXlsx)
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("open file %s fail: %s", path, err.Error())
		return nil, errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("读取模板文件失败：%s", err.Error()))
	}
	defer file.Close()

	msg := trpc.Message(ctx)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	head := thttp.Head(ctx)
	head.Response.Header().Set("Content-Type", "application/xlsx")
	head.Response.Header().Set("Content-Disposition", "attachment; filename="+"template.xlsx")
	_, _ = io.Copy(head.Response, file)
	return nil, nil
}

/***************** device group *****************/

// GetCollectDevicePoints /tbcm/device/points POST
func (c ConfigManager) GetCollectDevicePoints(ctx context.Context, req *pb.GetCollectDevicePointsReq) (*pb.GetCollectDevicePointsRsp, error) {
	rsp := &pb.GetCollectDevicePointsRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	id := req.GetId()
	gid, ok := cm.Worker().GetDeviceGidById(id)
	if !ok {
		return nil, errs.New(errcode.ErrCgiParamInvalid, "device not found")
	}
	if t, ok := cm.Worker().GetDeviceTemplateByGid(gid); ok {
		for _, info := range t.PointsInfo {
			var valueDef *structpb.Struct
			var err error
			if v, ok := info.ValueDef.(map[string]interface{}); ok {
				valueDef, err = utils.ConvertMapToStruct(v)
				if err != nil {
					return nil, errs.New(errcode.ErrCgiParamInvalid, rsp.Message)
				}
			} else {
				return nil, errs.New(errcode.ErrCgiDeviceValueDefErr, "template value def err")
			}
			//把第一个gid换成id
			//避免出现gid的数字在测点名中也有出现的情况（如gid为1，测点名为BatStatus_1）
			convertedId := strings.Replace(string(info.ID), string(gid), id, 1)
			point := pb.GetCollectDevicePointsRsp_CollectPoint{
				No:      info.ID.GetPointNo(),
				Id:      convertedId,
				Name:    info.Name,
				Access:  info.Access,
				Valdef:  valueDef,
				Valtype: info.ValueType,
			}
			rsp.Data = append(rsp.Data, &point)
		}
	}
	sort.Slice(rsp.Data, func(i, j int) bool {
		return rsp.Data[i].Id < rsp.Data[j].Id
	})
	return rsp, nil
}

// Control /tbcm/device/ctl POST
func (c ConfigManager) Control(ctx context.Context, req *pb.ControlReq) (*pb.StringDataRsp, error) {
	deviceId, pointId, err := definition.SplitDataPointID(definition.DataPointIDType(req.PointId))
	if err != nil {
		return nil, fmt.Errorf("invalid point id [%v], split failed", req.PointId)
	}

	// id to gid
	deviceGid, ok := cm.Worker().GetDeviceGidById(string(deviceId))
	if !ok {
		return nil, errs.New(errcode.ErrCgiParamInvalid, "device not found")
	}
	pointGid := string(deviceGid) + consts.DefaultIDSep + string(pointId)
	ctlInfo := model.PointControlInfo{
		DeviceId:  deviceId,
		DeviceGid: deviceGid,
		PointNo:   definition.DataPointIDType(req.PointId),
		PointGid:  definition.DataPointIDType(pointGid),
		Value:     req.Value,
	}

	if code, msg := dispatcher.Dispatcher().ControlPoint(ctlInfo); code != 0 {
		return &pb.StringDataRsp{
			Code:    int32(code),
			Message: msg,
			Data:    "",
		}, nil
	}
	return &pb.StringDataRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    errcode.DefaultCgiRspMessage,
	}, nil
}

// ProcessDevice /tbcm/device 设备配置操作，PUT编辑，DELETE删除
func (c ConfigManager) ProcessDevice(ctx context.Context, req *pb.ProcessDeviceReq) (*pb.StringDataRsp, error) {
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	rsp := &pb.StringDataRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	switch head.Request.Method {
	case "PUT":
		id := req.GetReqPut().GetId()

		device := req.GetReqPut().GetData()
		cmdInterval, err := strconv.Atoi(device.GetChannel().GetCmdIntetrval())
		if err != nil {
			cmdInterval = 0
		}
		waitTime, err := strconv.Atoi(device.GetChannel().GetWaitTime())
		if err != nil {
			waitTime = 0
		}
		timeout, err := strconv.Atoi(device.GetChannel().GetTimeout())
		if err != nil {
			timeout = 0
		}
		worker := cm.Worker()
		gid, ok := worker.GetDeviceGidById(id)
		if !ok {
			gid = worker.GetNextDeviceGid()
		}
		worker.SetDevice(definition.DeviceGidType(gid), model.Device{
			Gid:  definition.DeviceGidType(gid),
			Name: device.GetName(),
			ID:   device.GetId(),
			ChData: model.ChannelData{
				ChannelID:     device.GetChannel().GetChid(),
				ChannelParams: device.GetChannel().GetChparams(),
				Address:       device.GetChannel().GetAddr(),
				CmdInterval:   cmdInterval,
				WaitTimeMs:    waitTime,
				TimeoutMs:     timeout,
				Chtype:        device.GetChannel().GetChtype(),
			},
			TemplateData: model.TemplateInfo{
				TemplateName: device.GetTpl().GetTplnm(),
				TemplatePath: device.GetTpl().GetTplpath(),
			},
		})
		err = worker.SaveCurrentDevicesConfig(definition.CollectorDeviceTypeTBox)
		if err != nil {
			return nil, errs.New(errcode.ErrSaveConfigFail, fmt.Sprintf("save devices config fail: %v", err))
		}
		cm.NotifyConfigChange()
		rsp.Data = id
		return rsp, nil
	case "DELETE":
		id := req.GetReqDel().GetId()
		cm.Worker().DeleteDevicesByIds(id)
		rsp.Data = id
		return rsp, nil
	default:
		return nil, errs.New(
			errcode.ErrCgiHttpMethodNotSupported,
			fmt.Sprintf("method <%v> not support", head.Request.Method),
		)
	}
}

// GetCollectDevice /tbcm/device/fetch POST根据id获取设备配置
func (c ConfigManager) GetCollectDevice(ctx context.Context, req *pb.IDRequest) (*pb.GetDeviceRsp, error) {
	rsp := &pb.GetDeviceRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	id := req.GetId()
	device, ok := cm.Worker().GetDeviceById(id)
	if !ok {
		return nil, errs.New(errcode.ErrCgiParamInvalid, fmt.Sprintf("device id <%v> not exist", id))
	}
	rsp.Data = &pb.CollectDevice{
		Gid:  string(device.Gid),
		Id:   string(device.ID),
		Name: device.Name,
		Type: device.TypeEn,
		Tpl: &pb.Template{
			Tplnm:   device.TemplateData.TemplateName,
			Tplpath: device.TemplateData.TemplatePath,
		},
		Channel: &pb.Channel{
			Addr:         device.ChData.Address,
			Chid:         device.ChData.ChannelID,
			Chparams:     device.ChData.ChannelParams,
			Chtype:       device.ChData.Chtype,
			Timeout:      fmt.Sprintf("%v", device.ChData.TimeoutMs),
			CmdIntetrval: fmt.Sprintf("%v", device.ChData.CmdInterval),
			WaitTime:     fmt.Sprintf("%v", device.ChData.WaitTimeMs),
		},
	}
	return rsp, nil
}

// ExportAllTemplates /tbcm/tpls/excel GET导出所有模板
func (c ConfigManager) ExportAllTemplates(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if config.GetRB().Project.Source != "local" {
		return &emptypb.Empty{}, errs.New(errcode.ErrBadRequest, "当前不是local模式，无法执行该操作")
	}
	return nil, utils2.ExportAllTemplates(ctx)
}

// ExportTemplate /tbcm/tpl/excel/*/*/name 导出某个模板
func (c ConfigManager) ExportTemplate(ctx context.Context, req *pb.ExportTemplateRequest) (*emptypb.Empty, error) {
	if config.GetRB().Project.Source != "local" {
		return &emptypb.Empty{}, errs.New(errcode.ErrBadRequest, "当前不是local模式，无法执行该操作")
	}
	pathList := strings.Split(req.Path, "/")
	if len(pathList) != 3 {
		return nil, errs.New(errcode.ErrCgiParamInvalid, "请求路径须为 设备分类/厂商/名称 格式")
	}
	fileName := pathList[2]
	exportFileName, err := utils2.ExportTemplate(ctx, fileName)
	defer os.Remove(exportFileName)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	head := thttp.Head(ctx)
	if head == nil {
		return &emptypb.Empty{}, errs.New(errcode.ErrBadRequest, "head is nil")
	}
	msg := trpc.Message(ctx)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	resultF, err := os.Open(exportFileName)
	defer resultF.Close()
	if err != nil {
		return &emptypb.Empty{}, errs.New(errcode.ErrServerLogic, fmt.Sprintf("读取临时文件错误: %s", err.Error()))
	}
	contentType := "application/xlsx"
	head.Response.Header().Set("Content-Type", contentType)
	head.Response.Header().Set("Content-Disposition", "attachment; filename="+pathList[2]+".xlsx")
	_, err = io.Copy(head.Response, resultF)
	if err != nil {
		log.Errorf("write file content to http response err: %s, fileName: %s", err.Error(), fileName)
	}
	return &emptypb.Empty{}, nil
}

/***************** tpls group *****************/

// ProcessTemplates /tbcm/tpls GET获取模板列表,POST上传模板文件
func (c ConfigManager) ProcessTemplates(ctx context.Context, req *emptypb.Empty) (*pb.ProcessTemplatesRsp, error) {
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	rsp := &pb.ProcessTemplatesRsp{
		RspPost: &pb.StringDataRsp{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
			// Data:    "succeess",
		},
		RspGet: &pb.ProcessTemplatesRsp_RspGet{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
		},
	}
	switch head.Request.Method {
	case "GET":
		nodes := cm.Worker().GetTemplatesInfoNodes()
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})
		rsp.RspGet.Data = nodes
	case "POST":
		request := head.Request
		if err := request.ParseMultipartForm(100 * (1 << 20)); err != nil {
			return nil, errs.New(errcode.ErrCgiTemplateFileFail, "templates file build fail: parse multipartForm failed")
		}
		if request.MultipartForm == nil || request.MultipartForm.File == nil {
			return nil, errs.New(errcode.ErrCgiTemplateFileFail, "templates file build fail: parse multipartForm empty")
		}
		filesHeaders, ok := head.Request.MultipartForm.File["files"]
		if !ok {
			return nil, errs.New(errcode.ErrCgiTemplateFileFail, "templates file build fail: get files failed")
		}
		log.Info("TemplatesPost, get files ok")

		err := utils2.ProcessImportTemplate(filesHeaders)
		if err != nil {
			return nil, errs.New(errcode.ErrCgiTemplateFileFail, fmt.Sprintf("templates file build fail: %v", err.Error()))
		}
	default:
		return nil, errs.New(
			errcode.ErrCgiHttpMethodNotSupported,
			fmt.Sprintf("method <%v> not support", head.Request.Method),
		)
	}
	return rsp, nil
}

// ProcessTemplate /tbcm/tpl DELETE按模板路径删除模板 POST根据设备id获取模板信息
func (c ConfigManager) ProcessTemplate(ctx context.Context, req *pb.ProcessTemplateReq) (*pb.ProcessTemplateRsp, error) {
	// return nil, errs.New(errcode.ErrCgiNotImplemented, "not implemented")
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	rsp := &pb.ProcessTemplateRsp{
		RspDel: &pb.StringDataRsp{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
			// Data:    "succeess",
		},
		RspPost: &pb.Template{
			// Code:    errcode.DefaultCgiRspCode,
			// Message: errcode.DefaultCgiRspMessage,
		},
	}
	switch head.Request.Method {
	case "DELETE":
		// 格式应为 "设备分类/厂商/模板名称"
		//用户传入的路径，例如 "设备分类/厂商/模板名称"
		rawPath := req.GetReqDel().GetPath()
		if rawPath == "" {
			return nil, errs.New(errcode.ErrCgiParamInvalid, "path is empty")
		}

		//  检查是否包含路径穿越
		cleanPath := filepath.Clean(rawPath)
		if cleanPath != rawPath {
			// 如果 Clean 后结果不同，说明原始路径可能包含 ../ 或 ./ 等
			return nil, errs.New(errcode.ErrCgiParamInvalid, "invalid template path: contains path traversal")
		}

		// 按 "/" 分割路径假设前端传来的路径使用 "/" 分隔，如 "设备分类/厂商/模板名称"
		l := strings.Split(rawPath, "/")
		if len(l) != 3 {
			return nil, errs.New(errcode.ErrCgiParamInvalid, "template path error: expected '分类/厂商/模板名称'")
		}

		// 校验每个部分是否合法：非空、不包含 .. / \ 等非法字符
		for _, part := range l {
			if part == "" || part == "." || part == ".. " ||
				strings.Contains(part, "/") || strings.Contains(part, "\\") {
				return nil, errs.New(errcode.ErrCgiParamInvalid, "invalid path component: '"+part+"'")
			}
		}
		// 提取模板名称，此时已经确保安全
		templateName := l[2]

		// 删除模板配置
		err := cm.Worker().DeleteTemplateConfig(templateName)
		if err != nil {
			return nil, errs.New(errcode.ErrSaveConfigFail, "delete templates config fail")
		}
		cm.NotifyConfigChange()

		// 设置返回值
		rsp.RspDel.Data = rawPath
		// 若模板被删除，则可删除本地的excel文件
		// 安全本地 Excel 文件路径
		localExcelPath := filepath.Join(
			filepath.Clean(consts.ProjectPath),
			filepath.Clean(consts.RelativeExcelPath),
			filepath.Clean(consts.RelativeTemplateDir),
			templateName+".xlsx",
		)

		// 再次验证最终路径是否在预期的目录下，防止路径穿越
		baseDir := filepath.Join(
			filepath.Clean(consts.ProjectPath),
			filepath.Clean(consts.RelativeExcelPath),
			filepath.Clean(consts.RelativeTemplateDir),
		)
		//规范化路径
		cleanLocalExcelPath := filepath.Clean(localExcelPath)
		cleanBaseDir := filepath.Clean(baseDir)

		// 必须确保 localExcelPath 是 baseDir 的子目录下的文件，防止如 "../../etc/pass" 这类穿越
		if !strings.HasPrefix(cleanLocalExcelPath, cleanBaseDir+string(filepath.Separator)) {
			return nil, errs.New(errcode.ErrCgiParamInvalid, "invalid file path: potential path traversal detected")
		}

		// 执行删除操作
		err = os.Remove(cleanLocalExcelPath)
		if err != nil {
			if os.IsNotExist(err) {
				rsp.RspDel.Message = "file already deleted or not found"
			} else {
				return nil, errs.New(errcode.ErrCgiParamInvalid, "failed to remove excel file")
			}
		}
	default:
		return nil, errs.New(
			errcode.ErrCgiHttpMethodNotSupported,
			fmt.Sprintf("method <%v> not support", head.Request.Method),
		)
	}
	return rsp, nil
}

/***************** log group *****************/

// ProcessRunLog /tbcm/log/run POST根据指定的行数获取日志
func (c ConfigManager) ProcessRunLog(ctx context.Context, req *pb.ProcessRunLogReq) (*pb.ProcessRunLogRsp, error) {
	logContent, err := logfile.QueryServerLogFromFile("", int(req.GetNum()))
	if err != nil {
		return nil, errs.New(errcode.ErrCgiHandleFail, fmt.Sprintf("query log file failed, err: %v", err))
	}
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	switch head.Request.Method {
	case "GET":
		msg := trpc.Message(ctx)
		msg.WithSerializationType(codec.SerializationTypeUnsupported)
		head.Response.Header().Set("Content-Type", "text/plain")
		head.Response.Header().Set("Content-Disposition", "attachment; filename="+"run.log")
		_, _ = io.Copy(head.Response, strings.NewReader(logContent))
		return &pb.ProcessRunLogRsp{}, nil
	default:
		return &pb.ProcessRunLogRsp{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
			Data:    logContent,
		}, nil
	}
}

// ProcessLogPacket /tbcm/log/packet POST根据指定的行数获取日志
func (c ConfigManager) ProcessLogPacket(ctx context.Context, req *pb.ProcessLogPacketReq) (*pb.ProcessLogPacketRsp, error) {
	logPath := req.Chid
	if comConfig, ok := config.GetRB().Collector.Modbus.SerialsMap.COMs[req.Chid]; ok {
		logPath = comConfig.Dev
	}
	logName := logfile.GetPacketLogPath(logPath)
	logContent, _ := logfile.QueryLogFromFile(logName, int(req.GetNum()))
	// 兼容前端，传err后会导致前端页面不刷新
	//if err != nil {
	//	err = errs.New(errcode.ErrCgiHandleFail, fmt.Sprintf("query log file failed, err: %v", err))
	//}
	return &pb.ProcessLogPacketRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    logContent,
	}, nil
}

/***************** std group *****************/

// GetStdDevice /tbcm/std/device POST
func (c ConfigManager) GetStdDevice(ctx context.Context, req *emptypb.Empty) (*pb.GetStdDeviceRsp, error) {
	rsp := &pb.GetStdDeviceRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	device := cm.Worker().GetStdDeviceData()
	if device == nil {
		return rsp, nil
	}
	var rspList []*pb.GetStdDeviceRsp_Device
	for _, d := range device.StdDevices {
		rspList = append(rspList, &pb.GetStdDeviceRsp_Device{
			DeviceGid:    d.DeviceGid,
			DeviceNumber: d.DeviceNumber,

			DeviceNo:          d.DeviceNo,
			DeviceTypeEn:      d.DeviceTypeEn,
			DeviceTypeZh:      d.DeviceTypeZh,
			ApplicationTypeEn: d.ApplicationTypeEn,
			ApplicationTypeZh: d.ApplicationTypeZh,
		})
	}
	sort.Slice(rspList, func(i, j int) bool {
		return rspList[i].DeviceNumber < rspList[j].DeviceNumber
	})
	rsp.Data = rspList
	return rsp, nil
}

// copyStdDeviceTreeNode 递归复制树节点
func copyStdDeviceTreeNode(node *cm.StdDeviceTreeNode) *pb.GetStdDeviceTreeRsp_TreeNode {
	treeNode := &pb.GetStdDeviceTreeRsp_TreeNode{
		DeviceGid:         node.Device.DeviceGid,
		DeviceNumber:      node.Device.DeviceNumber,
		DeviceNumberShow:  node.Device.DeviceNumberShow,
		DeviceNo:          node.Device.DeviceNo,
		DeviceTypeEn:      node.Device.DeviceTypeEn,
		DeviceTypeZh:      node.Device.DeviceTypeZh,
		ApplicationTypeEn: node.Device.ApplicationTypeEn,
		ApplicationTypeZh: node.Device.ApplicationTypeZh,
		Children:          make([]*pb.GetStdDeviceTreeRsp_TreeNode, 0),
	}
	for _, child := range node.Children {
		treeNode.Children = append(treeNode.Children, copyStdDeviceTreeNode(child))
	}
	sort.Slice(treeNode.Children, func(i, j int) bool {
		return treeNode.Children[i].DeviceNumber < treeNode.Children[j].DeviceNumber
	})
	return treeNode
}

// GetStdDeviceTree /tbcm/std/device/tree GET
func (c ConfigManager) GetStdDeviceTree(ctx context.Context, req *emptypb.Empty) (*pb.GetStdDeviceTreeRsp, error) {
	rsp := &pb.GetStdDeviceTreeRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}

	stdDevicesTreeRoots := cm.Worker().GetStdDeviceTree()
	roots := []*pb.GetStdDeviceTreeRsp_TreeNode{}
	for _, node := range stdDevicesTreeRoots {
		roots = append(roots, copyStdDeviceTreeNode(node))
	}

	// Calculate device counts
	for _, root := range roots {
		root.DeviceCount = int32(len(root.Children))
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].DeviceNumber < roots[j].DeviceNumber
	})
	rsp.Data = roots
	return rsp, nil
}

// GetStdPoints /tbcm/std/points POST
func (c ConfigManager) GetStdPoints(ctx context.Context, req *pb.GetStdPointsReq) (*pb.GetStdPointsRsp, error) {
	rsp := &pb.GetStdPointsRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	gids := req.DeviceGid
	var rspList []*pb.DevicePointObject
	dataPoints := make(cmodel.DataPoints, 0, 0)

	stdDevice := cm.Worker().GetStdDeviceData()
	// 获取实时数据
	for _, gid := range gids {
		ps, ok := stdDevice.GetPointsByGid(gid)
		if !ok {
			continue
		}
		for _, p := range ps {
			rspList = append(rspList, &pb.DevicePointObject{
				DeviceGid:       p.StdDevice,
				PointKey:        fmt.Sprintf("%s.%s", p.StdDevice, p.StdPoint),
				DeviceNumber:    "",
				PointNameEn:     p.StdPoint,
				PointNameZh:     p.StdPointZh,
				Expression:      p.Expr,
				ExpressionMap:   p.Mapping,
				PointRw:         p.PointRw,
				PointLevel:      p.PointLevel,
				ValueType:       p.ValueType,
				Enable:          p.Enable,
				ValueValidRange: p.ValueValidRange,
				ValueUnit:       p.ValueUnit,
				ValuePrecision:  p.ValuePrecision,
				ValueEnum:       p.ValueEnum,
			})
			dataPoints = append(dataPoints, cmodel.DataPoint{
				ID: definition.DataPointIDType(fmt.Sprintf("%s.%s", p.StdDevice, p.StdPoint)),
			})
		}
	}
	rtdb.GetDataPoints(dataPoints)
	vMap := make(map[string]cmodel.DataPoint)
	for _, point := range dataPoints {
		vMap[string(point.ID)] = point
	}
	// 封装结果
	for _, p := range rspList {
		// 设置实时值
		pV, ok := vMap[p.PointKey]
		if !ok {
			continue
		}
		p.PointValue = pV.Rtd.Val.Pv.String()
		p.UpdateTime = fmt.Sprintf("%v", pV.Rtd.Val.Tms)
	}
	sort.Slice(rspList, func(i, j int) bool {
		return rspList[i].PointNameEn < rspList[j].PointNameEn
	})
	rsp.Data = rspList
	return rsp, nil
}

// ProcessStdSrc /tbcm/std/src 数据源查询、修改
func (c ConfigManager) ProcessStdSrc(ctx context.Context, req *pb.ProcessStdSrcReq) (*pb.ProcessStdSrcRsp, error) {
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}

	rsp := &pb.ProcessStdSrcRsp{
		RspPost: &pb.StdSrcPostRsp{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
			Data:    &pb.StdSrcData{},
		},
		RspPut: &pb.StdSrcPutRsp{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
		},
	}
	switch head.Request.Method {
	case "POST":
		request := req.GetReqPost()

		pk := request.GetPointKey()
		if pk == "" {
			return rsp, nil
		}
		// 获取标准测点的数据源
		ps := cm.Worker().GetStdDeviceData().GetPointsByPointKey(pk)
		if len(ps) == 0 {
			return rsp, nil
		}
		p := ps[0]
		paramMap, err := utils2.GetMappingObject(p.Mapping)
		if err != nil {
			return nil, errs.New(errcode.ErrCgiGetMappingFail, "get mapping fail:"+err.Error())
		}
		rsp.RspPost.Data.PointKey = pk
		rsp.RspPost.Data.Expression = p.Expr
		rsp.RspPost.Data.ExpressionMap = paramMap
		return rsp, nil
	case "PUT":
		// 修改
		pk := req.GetReqPut().GetPointKey()
		ep := req.GetReqPut().GetExpression()
		eMap := req.GetReqPut().GetExpressionMap()
		newPointList, err := cm.Worker().GetStdDeviceData().SavePointInfo(pk, ep, eMap)
		if err != nil {
			return nil, errs.New(errcode.ErrCgiSaveMappingFail, "save point fail:"+err.Error())
		}
		// 更新内存
		cm.Worker().SetStdPointData(newPointList)
		// 重新调度标准点计算
		if err := std.GetCalManager().Reload(); err != nil {
			log.Errorf("Reload std failed, %v", err)
			rsp.RspPut.Data = "Reload std failed"
		}
		// 保存文件 todo 保存后格式有变化
		//cm.Worker().SaveCurrentStdPointsConfig()
		rsp.RspPut.Data = "success"
		return rsp, nil
	default:
		return nil, errs.New(
			errcode.ErrCgiHttpMethodNotSupported,
			fmt.Sprintf("method <%v> not support", head.Request.Method),
		)
	}
}

// GetStdDevicesAndPoints /tbcm/standard/points 标准设备及测点查询
func (c ConfigManager) GetStdDevicesAndPoints(ctx context.Context, req *emptypb.Empty) (*pb.GetStdDevicesAndPointsRsp, error) {
	rsp := &pb.GetStdDevicesAndPointsRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    []*pb.GetStdDevicesAndPointsRsp_Device{},
	}
	stdDeviceData := cm.Worker().GetStdDeviceData().Copy()
	if stdDeviceData == nil {
		return &pb.GetStdDevicesAndPointsRsp{}, nil
	}

	roots := cm.Worker().GetStdDeviceTree()
	// 特殊处理逻辑
	// 对ITM，其子设备放到第一层
	for i := range roots {
		r := roots[i]
		if r.Device.ConciseCode == "ITM" {
			roots = append(roots, r.Children...)
			r.Children = []*cm.StdDeviceTreeNode{}
		}
	}
	for _, r := range roots {
		device := &pb.GetStdDevicesAndPointsRsp_Device{
			Id:      r.Device.ConciseCode,
			Type:    r.Device.DeviceTypeEn,
			Name:    r.Device.DeviceName,
			Points:  []*pb.GetStdDevicesAndPointsRsp_Point{},
			Devices: []*pb.GetStdDevicesAndPointsRsp_Device{},
		}
		// 处理测点
		ps, ok := stdDeviceData.GetPointsByGid(r.Device.DeviceGid)
		if ok {
			for _, p := range ps {
				device.Points = append(device.Points, &pb.GetStdDevicesAndPointsRsp_Point{
					Id:   r.Device.ConciseCode + "." + p.StdPoint,
					Name: p.StdPointZh,
				})
			}
			sort.Slice(device.Points, func(i, j int) bool {
				return device.Points[i].Id < device.Points[j].Id
			})
		}
		// 处理子设备及子设备的测点
		for _, sub := range r.Children {
			subDevice := &pb.GetStdDevicesAndPointsRsp_Device{
				Id:      sub.Device.ConciseCode,
				Type:    sub.Device.DeviceTypeEn,
				Name:    sub.Device.DeviceName,
				Points:  []*pb.GetStdDevicesAndPointsRsp_Point{},
				Devices: []*pb.GetStdDevicesAndPointsRsp_Device{},
			}

			ps, ok = stdDeviceData.GetPointsByGid(sub.Device.DeviceGid)
			if ok {
				for _, p := range ps {
					subDevice.Points = append(subDevice.Points, &pb.GetStdDevicesAndPointsRsp_Point{
						Id:   sub.Device.ConciseCode + "." + p.StdPoint,
						Name: p.StdPointZh,
					})
				}
				sort.Slice(subDevice.Points, func(i, j int) bool {
					return subDevice.Points[i].Id < subDevice.Points[j].Id
				})
			}
			device.Devices = append(device.Devices, subDevice)
		}
		sort.Slice(device.Devices, func(i, j int) bool {
			return device.Devices[i].Id < device.Devices[j].Id
		})
		rsp.Data = append(rsp.Data, device)
	}
	sort.Slice(rsp.Data, func(i, j int) bool {
		return rsp.Data[i].Id < rsp.Data[j].Id
	})
	return rsp, nil
}

/***************** network group *****************/

// ProcessNetwork /tbcm/network POST修改网络配置 GET获取当前网络配置
func (c ConfigManager) ProcessNetwork(ctx context.Context, req *pb.NetworkStatus) (*pb.ProcessNetworkRsp, error) {
	if config.GetRB().IsGatewayMode() {
		return nil, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}
	head := thttp.Head(ctx)
	if head == nil {
		return nil, errs.New(errcode.ErrNilHeader, "nil head")
	}
	switch head.Request.Method {
	case "POST":
		err := network.SetNetworkStatus(ctx, req)
		if err != nil {
			return nil, errs.New(errcode.ErrCgiHandleFail, err.Error())
		}
	case "GET":
		networkStatus, err := network.GetNetworkStatus(ctx)
		if err != nil {
			return nil, errs.New(errcode.ErrCgiHandleFail, err.Error())
		}
		return &pb.ProcessNetworkRsp{
			Code:    errcode.DefaultCgiRspCode,
			Message: errcode.DefaultCgiRspMessage,
			Data:    networkStatus,
		}, nil
	default:
		return nil, errs.New(
			errcode.ErrCgiHttpMethodNotSupported,
			fmt.Sprintf("method <%v> not support", head.Request.Method),
		)
	}
	rsp := &pb.ProcessNetworkRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
	}
	return rsp, nil
}

// EnableSwitch /tbcm/switch/enable
func (c ConfigManager) EnableSwitch(ctx context.Context, req *pb.EnableSwitchReq) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return nil, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}
	err := network.EnableSwitch(ctx, req)
	if err != nil {
		return nil, errs.New(errcode.ErrCgiHandleFail, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// DisableSwitch /tbcm/switch/disable
func (c ConfigManager) DisableSwitch(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return nil, errs.New(errcode.ErrGwMode, "agent-gw mode not support")
	}
	err := network.DisableSwitch(ctx)
	if err != nil {
		return nil, errs.New(errcode.ErrCgiHandleFail, err.Error())
	}
	return &emptypb.Empty{}, nil
}

/***************** other group *****************/

// Login /tbcm/login POST用户登录
func (c ConfigManager) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRsp, error) {
	token, err := usr.Login(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	expiration := time.Now().Add(30 * 24 * time.Hour)
	cookie := "nickname=" + req.Username + "; Path=/; Expires=" + expiration.Format(http.
		TimeFormat)
	thttp.Response(ctx).Header().Set("Set-Cookie", cookie)

	return &pb.LoginRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data: &pb.LoginRsp_LoginData{
			Token: token,
		},
	}, nil
}

// CmGroups /tbcm/groups GET
func (c ConfigManager) CmGroups(ctx context.Context, req *pb.CmGroupsReq) (*pb.CmGroupsRsp, error) {
	rsp := pb.CmGroupsRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data: []*pb.CmGroupsRsp_CmGroupsList{{
			Id:   "1",
			Name: "默认",
			Pid:  "-1",
		}},
	}
	return &rsp, nil
}

// GetSerials /tbcm/serials GET
func (c ConfigManager) GetSerials(ctx context.Context, req *emptypb.Empty) (*pb.GetSerialsRsp, error) {
	// serialsInfos := cm.GetSerialConfig()
	serialsInfos := config.Conf.GetSerialConfig()
	m := make(map[string]*pb.SerialInfo)
	for k, v := range serialsInfos {
		m[k] = &pb.SerialInfo{
			Baud:    v.Baud,
			Databit: v.Databit,
			Dev:     v.Dev,
			Id:      v.ID,
			Mode:    v.Mode,
			Parity:  v.Parity,
			Stopbit: v.Stopbit,
		}
	}
	return &pb.GetSerialsRsp{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    m,
	}, nil
}

// SetSerial /tbcm/serial PUT
func (c ConfigManager) SetSerial(ctx context.Context, req *pb.SetSerialReq) (*pb.SetSerialRsp, error) {
	return nil, errs.New(errcode.ErrCgiNotImplemented, "not implemented")
	// id := req.GetId()
	// data := req.GetData()
	// if data == nil {
	// 	return nil, errs.New(errcode.ErrCgiParamInvalid, "nil data")
	// }
	// serialConfig := &cm.SerialPortConfig{
	// 	Id:      id,
	// 	Dev:     data.GetDev(),
	// 	Baud:    data.GetBaud(),
	// 	Databit: data.GetDatabit(),
	// 	Stopbit: data.GetStopbit(),
	// 	Parity:  data.GetParity(),
	// 	Mode:    data.GetMode(),
	// }
	// cm.SetSerialConfig(id, serialConfig)
	// return &pb.SetSerialRsp{
	// 	Code:    errcode.DefaultCgiRspCode,
	// 	Message: errcode.DefaultCgiRspMessage,
	// 	Data:    id,
	// }, nil
}
