package repo

import (
	"etrpc-go/client/gorm"

	"alarm-compute/conf"
)

const (
	ALARM_MYSQL_NAME = "trpc.mysql.tbos.alarm"
)

// GetActiveFingerprints 获取当前活动告警的指纹
func GetActiveFingerprints(fingerprint []string) (list []string, err error) {
	if len(fingerprint) == 0 {
		return
	}
	ret := gorm.GetDB(
		ALARM_MYSQL_NAME).Table(
		conf.ServerConf.MysqlTableName.ActiveAlarmTable).Where(
		"fingerprint in (?)", fingerprint).Pluck("fingerprint", &list)
	if ret.Error != nil {
		err = ret.Error
		return
	}
	return
}
