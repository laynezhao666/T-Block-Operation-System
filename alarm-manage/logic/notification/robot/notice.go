package robot

import (
	"strconv"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/conf"
	"alarm-manage/entity/message"
	"alarm-manage/repo/rpc"
	cmodel "common/entity/model"
)

type RobotNoticeType int

const (
	AlarmRobotMsg RobotNoticeType = iota
	RestoreRobotMsg
)

// NoticeAlertByRobot NoticeAlertByRobot
func NoticeAlertByRobot(msgs []cmodel.AlarmActive) {
	for _, msg := range msgs {
		noticeByRobot(msg.ConvertToAlarmMsg(), int32(msg.MozuId), msg.AlarmID, AlarmRobotMsg)
	}
}

// NoticeRestoreByRobot NoticeRestoreByRobot
func NoticeRestoreByRobot(msgs []cmodel.AlarmHistory) {
	for _, msg := range msgs {
		noticeByRobot(msg.ConvertToRestoreMsg(), int32(msg.MozuId), msg.AlarmID, RestoreRobotMsg)
	}
}

// NoticeByRobot 通过群机器人下发告警（恢复）通知
func noticeByRobot(msg string, mozuId int32, alarmId int64, noticeType RobotNoticeType) {
	webhookList := conf.RobotConfig.Webhook[mozuId]
	for _, webhook := range webhookList {
		url := webhook.Url
		if url == "" {
			continue
		}
		switch noticeType {
		case AlarmRobotMsg:
			log.Debugf("NoticeByRobot Alarm")
			rpc.SendMarkdownButton(trpc.BackgroundContext(), url, msg, message.MarkdownAttachments{
				CallbackID: strconv.Itoa(int(alarmId)),
				Actions: []message.MarkdownAction{
					{
						Name:        "button1",
						Text:        "屏蔽此告警一小时",
						Type:        "button",
						Value:       "1",
						ReplaceText: "已经屏蔽此告警一小时",
					},
					{
						Name:        "button2",
						Text:        "屏蔽此告警24小时",
						Type:        "button",
						Value:       "24",
						ReplaceText: "已经屏蔽此告警24小时",
					},
				},
			})
		case RestoreRobotMsg:
			rpc.SendMarkdown(trpc.BackgroundContext(), url, msg)
		}
	}
}
