package tlink

import (
	"agent/entity/consts"
	"agent/entity/model"
	cmUtils "agent/repo/cm/utils"
)
// WriterImpl 实现
type WriterImpl struct {
	//collectorDeviceNum string
}
// NewWriterImpl 构造
func NewWriterImpl() *WriterImpl {
	w := &WriterImpl{
		//collectorDeviceNum: deviceNum,
	}
	return w
}
// SaveStdData 保存标准数据
func (w WriterImpl) SaveStdData(collectorDevice string, stdData *model.StdData) error {
	configMap := make(map[string]any, 1)
	configMap[collectorDevice] = stdData.StdPointsInfo
	err := cmUtils.SaveConfigMapToFile(
		consts.ProjectPath+"/"+collectorDevice+"/"+consts.RelativeStdFile,
		configMap,
	)
	if err != nil {
		return err
	}
	// 通知任务变化
	//cm.ConfigChangedChan() <- true
	return nil
}
// SaveDevices 保存设备信息
func (w WriterImpl) SaveDevices() ([]model.Device, error) {
	//TODO implement me
	panic("implement me")
}
// SaveTemplates 保存模板信息
func (w WriterImpl) SaveTemplates(list []string) (map[string]*model.TemplateData, error) {
	//TODO implement me
	panic("implement me")
}
