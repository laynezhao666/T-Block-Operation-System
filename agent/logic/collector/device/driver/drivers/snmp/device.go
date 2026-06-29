package snmp

import (
	"agent/entity/config"
	model2 "agent/entity/consts"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
	rtdbModel "agent/logic/collector/rtdb/model"
	"agent/utils"
	"agent/utils/osal"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"github.com/gosnmp/gosnmp"
)

// DeviceSNMP snmp设备
type DeviceSNMP struct {
	Data        model3.IDeviceData
	target      *gosnmp.GoSNMP
	isConnected bool
	TaskPool    *utils.TaskPool

	coroutinesEnable bool // 启用并发协程
	option           Option
}

func newGoSNMP(oidCount int, community string) *gosnmp.GoSNMP {
	g := gosnmp.Default
	var logger gosnmp.Logger
	if config.GetRB().IsFeatureEnable("snmp-log") {
		logger = gosnmp.NewLogger(utils.NewDefaultInfoLogger())
	}
	return &gosnmp.GoSNMP{
		Port:               g.Port,
		Transport:          g.Transport,
		Community:          community,
		Version:            g.Version,
		Timeout:            g.Timeout,
		Retries:            g.Retries,
		ExponentialTimeout: false,
		MaxOids:            oidCount,
		Logger:             logger,
		// 避免请求 IP 与响应 IP 不一致时无法正常采集
		UseUnconnectedUDPSocket: true,
	}
}

// Open 建立连接
func (s *DeviceSNMP) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) model2.Quality {
	s.option.Load(chanInfo, packets)

	args := strings.Split(chanInfo.Name, ":")
	if len(args) == 0 {
		return model2.QualityConfigError
	}
	s.target = newGoSNMP(s.option.PacketOIDs, s.option.ReadCommunity)
	s.target.Target = args[0]
	s.target.Port = defaultSnmpPort
	s.target.Timeout = time.Duration(s.option.ReadTimeOut) * time.Millisecond
	// CollectProtocolPacket 有连续失败一定次数才写 rtdb 的逻辑，默认置为 0
	s.target.Retries = s.option.ReadRetries

	if len(args) > 1 {
		p, err := strconv.ParseUint(args[1], 0, 16)
		if err != nil {
			return model2.QualityConfigError
		}
		s.target.Port = uint16(p)
	}

	if strings.Index(chanInfo.ProtocolVer, "V3") >= 0 {
		s.target = nil
	}
	if s.target == nil {
		return model2.QualityConfigError
	}

	driverCoroutines := s.option.ParallelCount
	if driverCoroutines > 1 {
		workers := make([]utils.TaskWorker, driverCoroutines)
		workers[0].Arg = s.target // 除第0个协程外，其它协程新创建GoSNMP库的连接对象
		for i := 1; i < driverCoroutines; i++ {
			snmp := newGoSNMP(s.option.PacketOIDs, s.option.ReadCommunity)
			snmp.Target = s.target.Target
			snmp.Port = s.target.Port
			snmp.Timeout = s.target.Timeout
			snmp.Retries = s.target.Retries
			workers[i].Arg = snmp
		}
		s.coroutinesEnable = true
		s.TaskPool = utils.NewTaskPool()
		s.TaskPool.Start(workers, s.taskHandle)
	}

	return model2.QualityOk
}

// Close 关闭连接
func (s *DeviceSNMP) Close() model2.Quality {
	if s.coroutinesEnable && s.TaskPool != nil {
		s.TaskPool.Stop()
	}

	if s.target != nil {
		if s.isConnected && s.target.Conn != nil {
			_ = s.target.Conn.Close()
		}
		s.isConnected = false
		s.target = nil
	}
	return model2.QualityOk
}

// Request 查询测点
func (s *DeviceSNMP) Request(ctx context.Context, packet *model.CollectProtocolPacket) (
	model2.Quality, model3.MessageStatistics) {
	var msgStat model3.MessageStatistics
	if packet == nil {
		return model2.QualityOk, msgStat
	}
	if !s.isConnected {
		if s.target == nil {
			return model2.QualityCannotOpen, msgStat
		}
		if err := s.target.Connect(); err != nil {
			log.Warnf("snmp driver connect error: %v", err)
			return model2.QualityCannotOpen, msgStat
		}
		s.isConnected = true
	}
	if s.coroutinesEnable {
		return s.requestByTask(ctx, packet)
	}
	pointsNum := len(packet.Points)
	requestNum := s.target.MaxOids
	index := 0
	oids := make([]string, requestNum)
	variants := make([]*osal.Variant, requestNum)
	points := make([]*model.PointInfo, requestNum)
	lastGetErrQuality := model2.QualityUncertain
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(s.option.TotalTimeOut)*time.Millisecond)
	defer cancel()
	for i := 0; i < pointsNum; i += requestNum {
		if utils.IsContextDone(ctx) {
			if errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
				log.Warnf("Request timeout, device:%v,host:%v", s.Data.Gid, s.target.Target)
			}
			return model2.QualityCmdRespTimeout, msgStat
		}
		if i+s.target.MaxOids > pointsNum {
			requestNum = pointsNum - i
		}
		oids = oids[:requestNum]
		variants = variants[:requestNum]
		points = points[:requestNum]
		for j := 0; j < requestNum; j++ {
			points[j] = packet.Points[index]
			valueParser, ok := packet.Points[index].Attr.ValParser.(*SnmpValueParser)
			if !ok || valueParser == nil {
				log.Warn("snmp value parser is not configured")
				return model2.QualityConfigError, msgStat
			}
			oids[j] = valueParser.OID
			variants[j] = &(points[j].RtVal.Pv)
			index++
		}
		quas := make([]model2.Quality, len(oids))
		setAllQua(quas, model2.QualityOk)
		msgStat.SendCount++
		ret := getValues(s.target, oids, variants, quas)
		if ret == model2.QualityOk {
			msgStat.SuccessCount++
		} else {
			lastGetErrQuality = ret
		}
		for j := 0; j < requestNum; j++ {
			// 当整个请求未出错，但单个测点出现错误时
			// 设置单个测点的 qua
			points[j].RtVal.Qua = ret
			if ret == model2.QualityOk && quas[j] != model2.QualityOk {
				points[j].RtVal.Qua = quas[j]
			}
			points[j].RtVal.Tms = utils.GetNowUTCTimeStamp()

			// snmp北向错误类型，需要设置qua
			handleNorthProtoValueErr(&points[j].RtVal)
		}
	}

	if msgStat.SuccessCount > 0 {
		return model2.QualityOk, msgStat
	}
	return lastGetErrQuality, msgStat
}

// 下层通讯管理机放在value中的北向错误类型（包含-9999、-99990--99999等），需要提取到q
func handleNorthProtoValueErr(val *rtdbModel.RTValue) {
	v, err := val.Pv.AsFloat()
	if err != nil {
		val.Qua = model2.QualityValueTypeError
		return
	}

	if utils.IsSpecialValue(v) {
		val.Qua = model2.QualityUnderBoxNorthErr
	}
}

// RequestPing 测试连接
func (s *DeviceSNMP) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) model2.Quality {
	testPacket := model.CollectProtocolPacket{
		Command: packet.Command,
	}
	if len(packet.Points) > s.target.MaxOids {
		testPacket.Points = packet.Points[:s.target.MaxOids]
	}

	ret, _ := s.Request(ctx, &testPacket)
	return ret
}

type requestTask struct {
	oids     []string
	variants []*osal.Variant
	points   []*model.PointInfo
	ret      model2.Quality
}

func newRequestTask(requestNum int) *requestTask {
	return &requestTask{
		oids:     make([]string, requestNum),
		variants: make([]*osal.Variant, requestNum),
		points:   make([]*model.PointInfo, requestNum),
	}
}

func (s *DeviceSNMP) requestByTask(ctx context.Context, packet *model.CollectProtocolPacket) (
	model2.Quality, model3.MessageStatistics) {
	var msgStat model3.MessageStatistics
	pointsNum := len(packet.Points)
	requestNum := s.target.MaxOids
	index := 0

	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(s.option.TotalTimeOut)*time.Millisecond)
	defer cancel()

	isContextTimeout := false
	tasks := make([]*requestTask, 0, pointsNum/requestNum+1)
	for i := 0; i < pointsNum; i += requestNum {
		if utils.IsContextDone(requestCtx) {
			if errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
				log.Warnf("requestByTask timeout, device:%v,host:%v", s.Data.Gid, s.target.Target)
			}
			isContextTimeout = true
			break
		}

		if i+s.target.MaxOids > pointsNum {
			requestNum = pointsNum - i
		}

		task := newRequestTask(requestNum)
		tasks = append(tasks, task)
		for j := 0; j < requestNum; j++ {
			task.points[j] = packet.Points[index]
			valueParser, ok := packet.Points[index].Attr.ValParser.(*SnmpValueParser)
			if !ok || valueParser == nil {
				log.Warn("snmp value parser is not configured")
				return model2.QualityConfigError, msgStat
			}
			task.oids[j] = valueParser.OID
			task.variants[j] = &(task.points[j].RtVal.Pv)
			index++
		}
		s.TaskPool.AddTask(task)
	}
	s.TaskPool.WaitFinish()

	lastGetErrQuality := model2.QualityOk
	for i := range tasks {
		msgStat.SendCount++
		if tasks[i].ret == model2.QualityOk {
			msgStat.SuccessCount++
		} else {
			lastGetErrQuality = tasks[i].ret
		}
	}

	if isContextTimeout {
		return model2.QualityCmdRespTimeout, msgStat
	}

	if msgStat.SuccessCount > 0 {
		return model2.QualityOk, msgStat
	}
	return lastGetErrQuality, msgStat
}

func (s *DeviceSNMP) taskHandle(task utils.Task, worker utils.TaskWorker) {
	reqTask := task.(*requestTask)
	target := worker.Arg.(*gosnmp.GoSNMP)
	var err error
	if target.Conn == nil {
		err = target.Connect()
	}

	quas := make([]model2.Quality, len(reqTask.oids))
	setAllQua(quas, model2.QualityOk)
	var ret model2.Quality
	if err != nil {
		ret = model2.QualityCannotOpen
	} else {
		ret = getValues(target, reqTask.oids, reqTask.variants, quas)
	}
	for j := 0; j < len(reqTask.oids); j++ {
		reqTask.points[j].RtVal.Qua = ret
		// 当整个请求未出错，但单个测点出现错误时
		// 设置单个测点的 qua
		if ret == model2.QualityOk && quas[j] != model2.QualityOk {
			reqTask.points[j].RtVal.Qua = quas[j]
		}
		reqTask.points[j].RtVal.Tms = utils.GetNowUTCTimeStamp()
	}
	reqTask.ret = ret
}

// Control 控制
func (s *DeviceSNMP) Control(packet *model.ControlProtocolPacket, val string) model2.Quality {
	if s == nil {
		return model2.QualityUncertain
	}
	valueParser, ok := packet.Point.Attr.ValParser.(*SnmpValueParser)
	if !ok {
		return model2.QualityUncertain
	}
	v := osal.NewVariantWithValue(val)
	return setValue(s.target, valueParser.OID, &v, valueParser.DataType)
}

func setAllQua(rets []model2.Quality, v model2.Quality) {
	for i := range rets {
		rets[i] = v
	}
}
