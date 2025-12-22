package kafka

import (
	"strings"

	"agent/entity/consts"
	"agent/entity/model/data"
)



func setErrorStatus(data *data.DataUnit, distributeError error) {
	// todo
}

func isIDCDevice(deviceID string) bool {
	return strings.HasPrefix(deviceID, consts.IDCDevicePrefix)
}