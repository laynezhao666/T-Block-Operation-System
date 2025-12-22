// Package build 采集相关配置导入
package build

import (
	"cmdb/entity/cond"
	"cmdb/repo/db"
	"common/entity/model"
	"context"
	"etrpc-go/util/copyutil"
	"fmt"
	"github.com/samber/lo"
	"sync"
	"trpc.group/trpc-go/trpc-go"
	"trpcprotocol/cmdb"
)

// IConfigBuildApi 采集配置导入相关接口
type IConfigBuildApi interface {
	// ImportModel 导入模型数据
	ImportModel(ctx context.Context, req *cmdb.ReqImportModel) (*cmdb.RspImportModel, error)
	// ListMozu 获取模组列表
	ListMozu(ctx context.Context, req *cmdb.ReqListMozu) (*cmdb.RspListMozu, error)
	// SaveMozu 保存模组
	SaveMozu(ctx context.Context, req *cmdb.ReqSaveMozu) error
	// DeleteMozu 删除模组
	DeleteMozu(ctx context.Context, req *cmdb.ReqDeleteMozu) error
}

var (
	buildApi IConfigBuildApi
	initOnce sync.Once
)

// GetConfigImportApi 创建采集配置导入相关实现类
func GetConfigImportApi() IConfigBuildApi {
	initOnce.Do(func() {
		// init config build api object
		obj := &configBuildApiImpl{
			deviceEntityDao:           db.NewDeviceEntityDao(),
			devicePointDao:            db.NewDevicePointDao(),
			collectorDeviceDao:        db.NewCollectorDeviceDao(),
			collectorTemplateDao:      db.NewCollectorTemplateDao(),
			collectorTemplatePointDao: db.NewCollectorTemplatePointDao(),
			mozuInfoDao:               db.NewMozuInfoDao(),
			alarmStrategyDao:          db.NewAlarmStrategyDao(),
		}
		buildApi = obj
	})
	return buildApi
}

type configBuildApiImpl struct {
	deviceEntityDao           db.IDeviceEntityDao
	devicePointDao            db.IDevicePointDao
	collectorDeviceDao        db.ICollectorDeviceDao
	collectorTemplateDao      db.ICollectorTemplateDao
	collectorTemplatePointDao db.ICollectorTemplatePointDao
	mozuInfoDao               db.IMozuInfoDao
	alarmStrategyDao          db.IAlarmStrategyDao
}

// SaveMozu 保存模组
func (obj *configBuildApiImpl) SaveMozu(ctx context.Context, req *cmdb.ReqSaveMozu) error {
	newMozuInfo := &model.MozuInfo{}
	_ = copyutil.Copy(req, newMozuInfo)
	if err := obj.mozuInfoDao.Save(ctx, newMozuInfo); err != nil {
		return err
	}
	return nil
}

// ListMozu 获取模组列表
func (obj *configBuildApiImpl) ListMozu(ctx context.Context, req *cmdb.ReqListMozu) (*cmdb.RspListMozu, error) {
	list, _, err := obj.mozuInfoDao.List(ctx, &cond.ListMozuInfoCond{
		MozuId:   req.MozuId,
		MozuName: req.MozuName,
		MozuCode: req.MozuCode,
	})
	if err != nil {
		return nil, err
	}
	// 转换成返回格式
	rspList := lo.Map(list, func(item *model.MozuInfo, index int) *cmdb.RspListMozu_MozuInfo {
		rspItem := &cmdb.RspListMozu_MozuInfo{}
		_ = copyutil.Copy(item, &rspItem)
		return rspItem
	})
	return &cmdb.RspListMozu{List: rspList}, nil
}

func (obj *configBuildApiImpl) DeleteMozu(ctx context.Context, req *cmdb.ReqDeleteMozu) error {
	return obj.mozuInfoDao.Delete(ctx, req.MozuId)
}

// ImportModel 导入模型数据
func (obj *configBuildApiImpl) ImportModel(ctx context.Context, req *cmdb.ReqImportModel) (*cmdb.RspImportModel, error) {
	list, _, err := obj.mozuInfoDao.List(ctx, &cond.ListMozuInfoCond{
		MozuId: []int32{req.MozuId},
	})
	if err != nil {
		return nil, err
	}
	if len(list) != 1 {
		return nil, fmt.Errorf("模组不存在,mozu_id:[%d]", req.MozuId)
	}

	var allErrs []string

	// 解析标准设备文件
	if req.DeviceEntity != nil {

		entities, errs, err := parseDeviceEntityExcel(req, "标准设备文件")
		if err != nil {
			allErrs = append(allErrs, "标准设备导入失败: "+err.Error())
		}
		allErrs = append(allErrs, errs...)
		mozuName := list[0].MozuName
		for _, item := range entities {
			item.MozuName = mozuName
		}

		// 保存到数据库
		if len(entities) > 0 {
			err := obj.deviceEntityDao.BatchUpdate(ctx, req.MozuId, entities)
			if err != nil {
				allErrs = append(allErrs, "标准设备保存失败: "+err.Error())
			}
		}
	}

	msg := trpc.Message(ctx)
	msg.ClientReqHead()

	// 解析标准测点文件
	if req.DevicePoint != nil {
		points, errs, err := parseDevicePointExcel(req, "标准测点文件")
		if err != nil {
			allErrs = append(allErrs, "标准测点导入失败: "+err.Error())
		}
		allErrs = append(allErrs, errs...)

		// 保存到数据库
		if len(points) > 0 {
			err := obj.devicePointDao.BatchUpdate(ctx, req.MozuId, points)
			if err != nil {
				allErrs = append(allErrs, "标准测点保存失败: "+err.Error())
			}
		}
	}

	// 解析采集设备文件
	if req.CollectorDevice != nil {
		devices, errs, err := parseCollectorDeviceExcel(req, "采集设备文件")
		if err != nil {
			allErrs = append(allErrs, "采集设备导入失败: "+err.Error())
		}
		allErrs = append(allErrs, errs...)

		// 保存到数据库
		if len(devices) > 0 {
			err := obj.collectorDeviceDao.BatchUpdate(ctx, req.MozuId, devices)
			if err != nil {
				allErrs = append(allErrs, "采集设备保存失败: "+err.Error())
			}
		}
	}

	// 解析采集模版文件
	if req.CollectorTemplate != nil {
		templates, errs, err := parseCollectorTemplateExcel(req, "采集模版文件")
		if err != nil {
			allErrs = append(allErrs, "采集模版导入失败: "+err.Error())
		}
		allErrs = append(allErrs, errs...)

		// 保存到数据库
		if len(templates) > 0 {
			err := obj.collectorTemplateDao.BatchUpdate(ctx, req.MozuId, templates)
			if err != nil {
				allErrs = append(allErrs, "采集模版保存失败: "+err.Error())
			}
		}
	}

	// 解析采集模版测点文件
	if req.TemplatePoint != nil {
		points, errs, err := parseCollectorTemplatePointExcel(req, "采集模版测点文件")
		if err != nil {
			allErrs = append(allErrs, "采集模版测点导入失败: "+err.Error())
		}
		allErrs = append(allErrs, errs...)

		// 保存到数据库
		if len(points) > 0 {
			err := obj.collectorTemplatePointDao.BatchUpdate(ctx, req.MozuId, points)
			if err != nil {
				allErrs = append(allErrs, "采集模版测点保存失败: "+err.Error())
			}
		}
	}

	// 解析告警策略文件
	if req.AlarmStrategy != nil {
		strategies, errs, err := parseAlarmStrategyExcel(req, "告警策略文件")
		if err != nil {
			allErrs = append(allErrs, "告警策略导入失败: "+err.Error())
		}
		allErrs = append(allErrs, errs...)

		// 保存到数据库
		if len(strategies) > 0 {
			err := obj.alarmStrategyDao.BatchUpdate(ctx, req.MozuId, strategies)
			if err != nil {
				allErrs = append(allErrs, "告警策略保存失败: "+err.Error())
			}
		}
	}

	updateMozuInfo := &model.MozuInfo{
		MozuId:         req.MozuId,
		PublishVersion: req.Version,
	}
	if err := obj.mozuInfoDao.Save(ctx, updateMozuInfo); err != nil {
		allErrs = append(allErrs, "mozu信息保存失败: "+err.Error())
	}

	return &cmdb.RspImportModel{Errs: allErrs}, nil
}
