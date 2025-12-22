// Package notification notification
package notification

import (
	"sync"

	"github.com/samber/lo"

	"alarm-manage/conf"
	"alarm-manage/repo/rpc"
	cmodel "common/entity/model"
)

var (
	speaker *speakerImpl
	once    sync.Once
)

// ISpeakerApi 通知接口
type ISpeakerApi interface {
	ReportAlert(alertList []cmodel.AlarmActive) error
	BatchReportAlert(alertList []cmodel.AlarmActive) error
	ReportRestore(restoreList []cmodel.AlarmHistory) error
}

type speakerImpl struct {
}

// GetSpeaker 获取通知接口
func GetSpeaker() ISpeakerApi {
	once.Do(func() {
		speaker = &speakerImpl{}
	})
	return speaker
}

// ReportAlert 批量上报活动告警
// 1. 发送cgi服务，本地动环推送
// 2. 发送foc云端kafka
// 3. 发送机器人
func (s *speakerImpl) ReportAlert(alertList []cmodel.AlarmActive) error {
	if len(alertList) == 0 {
		return nil
	}
	rpc.GetCkafka().SendCgiAlarm(alertList)
	// robot.NoticeAlertByRobot(alertList)
	return nil
}

// ReportRestore 上报恢复告警
func (s *speakerImpl) ReportRestore(restoreList []cmodel.AlarmHistory) error {
	if len(restoreList) == 0 {
		return nil
	}
	// robot.NoticeRestoreByRobot(restoreList)
	return nil
}

// BatchReportAlert 批量上报活动告警
// 单独用于接口调用，发送全量活动告警
func (s *speakerImpl) BatchReportAlert(alertList []cmodel.AlarmActive) error {
	if len(alertList) == 0 {
		return nil
	}
	batchSize := conf.ServerConf.AlertManageConfig.BatchChannelSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	chunkList := lo.Chunk(alertList, int(batchSize))
	for _, chunk := range chunkList {
		err := rpc.GetCkafka().SendCgiAlarm(chunk)
		if err != nil {
			return err
		}
	}
	return nil
}
