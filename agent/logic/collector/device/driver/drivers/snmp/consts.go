package snmp

import (
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"

	"agent/utils/flog"
)

var (
	needRetry = false
	filterLog *flog.Filter
)

// Init constants
func Init() error {
	needRetry = config.GetRB().IsFeatureEnable("snmp-retry")

	snmpLogTime := config.LoadIntOrDefault(config.GetRB().Collector.Snmp.LogInterval, 10)
	filterLog = flog.NewFilterLogger(time.Duration(snmpLogTime)*time.Minute, log.GetDefaultLogger())

	return nil
}
