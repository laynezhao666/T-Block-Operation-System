package rtask

import (
	"fmt"
	"strconv"
	"time"

	"etrpc-go/log"

	"alarm-compute/conf"
	"alarm-compute/entity/epoint"
	"alarm-compute/entity/taskcode"
	"alarm-compute/logic/collector/vtpoint"
	"alarm-compute/utils/common"
	"alarm-compute/utils/modcall"
)

const (
	VirtualPointTemplate = "%s.%s"
)

// StartVirtualRuleTask 虚拟测点计算任务
func (rt *RuleTask) StartVirtualRuleTask(pointValue epoint.HistoryValueMap, t time.Time) error {
	startTime := time.Now()
	defer func() {
		modcall.RecordAlarmComputeTime(rt.ServiceName(), float64(time.Since(startTime).Milliseconds()))
	}()
	_, err := rt.CheckMissDelayPointList(t.Unix(), pointValue, true)
	if err != nil {
		log.Warnf("virtual task not running, CheckMissPointList failed, key: %s, err: %v", rt.GetKey(), err)
		return &taskcode.PointDataLackErr
	}
	expRes, err := rt.Alert.StartVirtualAnalyze(t.Unix(), pointValue)
	if err != nil {
		// 失败处理
		log.Warnf("vt alert StartVirtualRT failed, expr: %v, key: %s, err: %v",
			common.JSONMarshalNoErr(rt.Alert.Exp.Express), rt.GetKey(), err)
		return &taskcode.ExprAnalyzeErr
	}
	pointName := fmt.Sprintf(VirtualPointTemplate, rt.Gid, rt.AlarmName)
	// TODO 发送数据Kafka
	// log.Infof("pointName %s at time %d is %f", pointName, t.Unix(), expRes)
	rt.SendVtMsg2Collector(pointName, t, expRes)
	return nil
}

// SendVtMsg2Collector 发送虚拟测点消息
func (rt *RuleTask) SendVtMsg2Collector(pointName string, t time.Time, value float64) {
	newPointMsg := &vtpoint.VirtualPoint{
		I: pointName,
		V: strconv.FormatFloat(value, 'f', int(conf.ServerConf.VirtualConfig.RoundPrecision), 64),
		Q: "0",
		T: strconv.FormatInt(t.Unix(), 10),
	}
	vtpoint.GetPointCollector().AddVirtualPointData(newPointMsg)
}
