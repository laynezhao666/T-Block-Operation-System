package opc

import (
	"context"
	"fmt"
	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/cm"
	"agent/logic/collector/device/model"
	rtdbModel "agent/logic/collector/rtdb/model"
	"agent/utils"
	"agent/utils/flog"
	"agent/utils/osal"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

const (
	enableSubMod       = true
	subscribeBatchSize = 2000
	defaultSubInterval = time.Second
)

var (
	filterLog *flog.Filter
)

const (
	// protocolVerV2 v2协议版本，reg字段为完整的OPC监控项路径，直接用于构造NodeID
	protocolVerV2 = "v2"
)

// Device 管理 OPC UA 客户端与订阅（每个 Device 自带独立 cache,避免cache过期reopen时相互影响）
type Device struct {
	client    *opcua.Client
	notifyChs []chan *opcua.PublishNotificationData
	ctx       context.Context
	cancel    context.CancelFunc

	subs        []*opcua.Subscription
	cache       *dataCache // per-device cache
	protocolVer string     // 协议版本，来自 drvinfo.protver
}

// Open 打开驱动并按点订阅
func (d *Device) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	filterLog = flog.NewFilterLogger(time.Duration(10)*time.Minute, log.GetDefaultLogger())

	log.Warnf("[OPCUA] Open: channel=%v", chanInfo)

	// 保存协议版本，用于区分不同的NodeID构建方式
	d.protocolVer = chanInfo.ProtocolVer

	// 每个 device 初始化独立 cache
	d.cache = NewDataCache()

	d.ctx, d.cancel = context.WithCancel(context.Background())
	// 构造 endpoint URL：opc.tcp://{Name}{Address}
	// Name 为 IP:Port（如 10.123.196.246:12686）
	// Address 为 URL 路径部分（如 /milo），可选
	endpoint := "opc.tcp://" + chanInfo.Name
	if chanInfo.Address != "" {
		if !strings.HasPrefix(chanInfo.Address, "/") {
			endpoint += "/"
		}
		endpoint += chanInfo.Address
	}
	log.Warnf("[OPCUA] endpoint URL: %v", endpoint)

	// 发现端点
	endpoints, err := opcua.GetEndpoints(d.ctx, endpoint)
	if err != nil {
		log.Errorf("[OPCUA] GetEndpoints fail: %v", err)
		return consts.QualityCannotOpen
	}
	policy := "None"
	mode := "None"
	ep, err := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	if err != nil {
		log.Errorf("[OPCUA] SelectEndpoint fail:%v", err)
		return consts.QualityCannotOpen
	}
	ep.EndpointURL = endpoint
	log.Warnf("[OPCUA] Selected endpoint: SecurityPolicy=%v SecurityMode=%v", ep.SecurityPolicyURI, ep.SecurityMode)

	// 解析认证信息：Params 格式为 "username:password"，为空则匿名登录
	var authOpts []opcua.Option
	if chanInfo.Params != "" {
		parts := strings.SplitN(chanInfo.Params, ":", 2)
		if len(parts) == 2 {
			username, password := parts[0], parts[1]
			log.Warnf("[OPCUA] using username/password auth, user=%v", username)
			authOpts = []opcua.Option{
				opcua.AuthUsername(username, password),
				opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeUserName),
			}
		} else {
			log.Warnf("[OPCUA] invalid Params format [%v], fallback to anonymous", chanInfo.Params)
			authOpts = []opcua.Option{
				opcua.AuthAnonymous(),
				opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
			}
		}
	} else {
		authOpts = []opcua.Option{
			opcua.AuthAnonymous(),
			opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
		}
	}

	// 建立连接
	opts := []opcua.Option{
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
		opcua.CertificateFile(""),
		opcua.PrivateKeyFile(""),
	}
	opts = append(opts, authOpts...)
	d.client, err = opcua.NewClient(ep.EndpointURL, opts...)
	if err != nil {
		log.Errorf("[OPCUA] NewClient fail:%v", err)
		return consts.QualityCannotOpen
	}
	if err := d.client.Connect(d.ctx); err != nil {
		log.Errorf("[OPCUA] Connect fail:%v", err)
		return consts.QualityCannotOpen
	}

	// 构建监控项
	miCreateRequests := d.CreateMonitorItems(packets)
	if miCreateRequests == nil {
		log.Errorf("[OPCUA] CreateMonitorItems nil, subscribe aborted")
		return consts.QualitySubscribeFail
	}

	if enableSubMod {
		if err := d.createSubscriptions(miCreateRequests); err != nil {
			log.Errorf("[OPCUA] initial subscribe failed: %v", err)
			return consts.QualitySubscribeFail
		}
	} else {
		log.Warnf("[OPCUA] enable read mode (no subscription)")
	}

	return consts.QualityOk
}

// CreateMonitorItems 构造监控项请求（根据现有设备/点位映射）
func (d *Device) CreateMonitorItems(packets model.ListCollectPackets) []*ua.MonitoredItemCreateRequest {
	var miCreateRequests []*ua.MonitoredItemCreateRequest

	// 构造实际订阅名称：根据gid 到 设备编号的对应关系
	subGid2Id := cm.Worker().GetMapSubDevicesGid2Id()
	for _, packet := range packets {
		for _, point := range packet.Points {
			deviceGid, pointId, err := definition.SplitDataPointID(point.Attr.ID)
			if err != nil {
				log.Errorf("[OPCUA] invalid point id [%v], split failed", point.Attr.ID)
				continue
			}
			deviceId, ok := subGid2Id[deviceGid]
			if !ok {
				log.Errorf("[OPCUA] deviceGid not exist: %+v", point)
				continue
			}

			// 尝试从 ValParser 中获取 reg（即 protdef.reg / val_key）值
			// 如果 reg 与 pointId 不同，说明 reg 是完整的 OPC 监控项路径，直接使用
			nodeID := d.buildNodeID(point, deviceId, string(pointId))

			req, err := d.cache.GenMonitoredItemRequest(nodeID, string(point.Attr.ID))
			if err != nil {
				log.Errorf("[OPCUA] GenMonitoredItemRequest fail: %v", err)
				return nil
			}
			miCreateRequests = append(miCreateRequests, req)
		}
	}
	log.Warnf("[OPCUA] CreateMonitorItems count=%v", len(miCreateRequests))
	return miCreateRequests
}

// buildNodeID 构建 OPC UA NodeID
// v2 协议：reg 字段为完整的 OPC 监控项路径，使用 ns=2;s={reg} 构造 NodeID
// 默认协议：按常规方式拼接 ns=1;s=t|{deviceId}.{pointId}
func (d *Device) buildNodeID(point *model.PointInfo, deviceId string, pointId string) string {
	// v2 协议：直接使用 reg（ValParser.Addr）作为 NodeID 路径，namespace=2
	if d.protocolVer == protocolVerV2 {
		if vp, ok := point.Attr.ValParser.(*ValueParser); ok && vp != nil && vp.Addr != "" {
			nodeID := fmt.Sprintf("ns=2;s=%v", vp.Addr)
			log.Debugf("[OPCUA] v2 use reg as nodeID: %v", nodeID)
			return nodeID
		}
	}

	// 默认协议：使用 sub_device + pointId 拼接
	if config.GetRB().IsOpcSpecialId() {
		deviceId = strings.Replace(deviceId, "-", "_", -1)
	}
	return fmt.Sprintf("ns=1;s=t|%v%v%v", deviceId, consts.DefaultIDSep, pointId)
}

// createSubscriptions 依据请求列表分批创建订阅与监控项
func (d *Device) createSubscriptions(miCreateRequests []*ua.MonitoredItemCreateRequest) error {
	// 清理旧订阅
	if len(d.subs) > 0 || len(d.notifyChs) > 0 {
		log.Warnf("[OPCUA] cleanup old subscriptions: subs=%d, chans=%d", len(d.subs), len(d.notifyChs))
		for _, sub := range d.subs {
			if sub != nil && d.ctx != nil {
				_ = sub.Cancel(d.ctx)
			}
		}
		for _, ch := range d.notifyChs {
			if ch != nil {
				close(ch)
			}
		}
		d.subs = nil
		d.notifyChs = nil
	}

	// 分批订阅
	for i := 0; i < len(miCreateRequests); i += subscribeBatchSize {
		end := i + subscribeBatchSize
		if end > len(miCreateRequests) {
			end = len(miCreateRequests)
		}
		batch := miCreateRequests[i:end]
		successCount := 0

		notifyCh := make(chan *opcua.PublishNotificationData)
		d.notifyChs = append(d.notifyChs, notifyCh)

		sub, err := d.client.Subscribe(d.ctx, &opcua.SubscriptionParameters{Interval: defaultSubInterval}, notifyCh)
		if err != nil {
			log.Errorf("[OPCUA] Subscribe fail:%v", err)
			return err
		}
		log.Warnf("[OPCUA] Created subscription id=%v, batch size=%v", sub.SubscriptionID, len(batch))
		d.subs = append(d.subs, sub)

		res, err := sub.Monitor(d.ctx, ua.TimestampsToReturnBoth, batch...)
		if err != nil {
			log.Errorf("[OPCUA] Monitor fail: %v", err)
			return err
		}
		for j, r := range res.Results {
			if r.StatusCode != ua.StatusOK {
				nodeID := batch[j].ItemToMonitor.NodeID.String()
				log.Debugf("[OPCUA] Monitor item fail, nodeID=%s status=%v", nodeID, r.StatusCode)
			} else {
				successCount++
			}
		}

		// 启动消费 loop
		go d.loop(notifyCh)
		log.Warnf("[OPCUA] Subscribed ok=%v fail=%v total=%v (batch %d~%d)", successCount, len(batch)-successCount, len(batch), i, end-1)
		time.Sleep(time.Second)
	}
	// 订阅建立成功，重置推送计数器并刷新 lastOkTime
	d.cache.ResetPushCount()
	d.cache.SetLastOk()
	return nil
}

// scheduleResubscribe f
func (d *Device) scheduleResubscribe(reason string) {
	// 快照 本 device 的 (reportId, nodeID) 重新建请求
	pairs := d.cache.ListAllReportNodePairs()
	if len(pairs) == 0 {
		log.Warnf("[OPCUA] resubscribe skipped: no monitored items to rebuild")
		return
	}
	reqs := make([]*ua.MonitoredItemCreateRequest, 0, len(pairs))
	for _, p := range pairs {
		rid, nid := p[0], p[1]
		req, err := d.cache.GenMonitoredItemRequest(nid, rid)
		if err != nil {
			log.Errorf("[OPCUA] rebuild GenMonitoredItemRequest fail: %v", err)
			return
		}
		reqs = append(reqs, req)
	}

	if err := d.createSubscriptions(reqs); err != nil {
		log.Errorf("[OPCUA] resubscribe failed once: %v", err)
		return
	}
	log.Warnf("[OPCUA] resubscribe success, items=%d", len(reqs))
}

// loop 消费订阅通知：订阅级异常 → 清缓存 + 触发一次性重建
func (d *Device) loop(notifyCh chan *opcua.PublishNotificationData) {
	for {
		select {
		case <-d.ctx.Done():
			log.Warnf("[OPCUA] loop exit: context done")
			return
		case res, ok := <-notifyCh:
			if !ok {
				log.Warn("[OPCUA] notifyCh closed, exit loop")
				return
			}
			if res == nil {
				log.Warn("[OPCUA] nil PublishNotificationData")
				continue
			}

			// 订阅级错误 → 清空本 device 的订阅缓存并触发重建
			if res.Error != nil {
				log.Errorf("[OPCUA] publish error: %v, invalidate all cache and resubscribe once", res.Error)
				d.cache.InvalidateAllForSubscription()
				d.scheduleResubscribe("publish error")
				return
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				// keep-alive: 无监控项，即链路活着
				if len(x.MonitoredItems) == 0 {
					log.Warnf("[OPCUA] keep-alive received")
					d.cache.SetLastOk()
					return
				}

				anyGood := false
				for _, item := range x.MonitoredItems {
					if item.Value.Status == ua.StatusOK || item.Value.Status == ua.StatusGood {
						anyGood = true
						// 正常值入库...
						tms := item.Value.SourceTimestamp.Unix()
						if item.Value.SourceTimestamp.IsZero() {
							tms = utils.GetNowUTCTimeStamp()
						}
						rtValue := rtdbModel.RTValue{
							Pv:  osal.NewVariantWithValue(VariantToGoValue(item.Value.Value)),
							Qua: UaStatusCode2Quality(item.Value.Status),
							Tms: tms,
						}
						if ok := d.cache.SetPointValue(item.ClientHandle, rtValue); !ok {
							log.Warnf("[OPCUA] SetPointValue miss handle=%v", item.ClientHandle)
						}
					} else {
						// 非 Good：清理该点缓存
						if ok := d.cache.InvalidateByHandle(item.ClientHandle); !ok {
							log.Warnf("[OPCUA] InvalidateByHandle miss, handle=%v", item.ClientHandle)
						}
					}
				}
				if anyGood {
					d.cache.SetLastOk()
				}

			case *ua.StatusChangeNotification:
				log.Errorf("[OPCUA] status change: %v, invalidate all cache and resubscribe once", x.Status)
				d.cache.InvalidateAllForSubscription()
				d.scheduleResubscribe("status change")
				return
			default:
				log.Debugf("[OPCUA] publish type: %T", res.Value)
			}
		}
	}
}

// VariantToGoValue 将 *ua.Variant 转 Go 基础类型/切片
func VariantToGoValue(v *ua.Variant) interface{} {
	if v == nil {
		return nil
	}
	switch v.Value().(type) {
	case nil:
		return nil
	case bool:
		if v.Value().(bool) {
			return 1
		}
		return 0
	case int8:
		return v.Value().(int8)
	case uint8:
		return v.Value().(uint8)
	case int16:
		return v.Value().(int16)
	case uint16:
		return v.Value().(uint16)
	case int32:
		return v.Value().(int32)
	case uint32:
		return v.Value().(uint32)
	case int64:
		return v.Value().(int64)
	case uint64:
		return v.Value().(uint64)
	case float32:
		return v.Value().(float32)
	case float64:
		return v.Value().(float64)
	case string:
		return v.Value().(string)
	case time.Time:
		return v.Value().(time.Time)
	case []bool:
		return v.Value().([]bool)
	case []int8:
		return v.Value().([]int8)
	case []uint8:
		return v.Value().([]uint8)
	case []int16:
		return v.Value().([]int16)
	case []uint16:
		return v.Value().([]uint16)
	case []int32:
		return v.Value().([]int32)
	case []uint32:
		return v.Value().([]uint32)
	case []int64:
		return v.Value().([]int64)
	case []uint64:
		return v.Value().([]uint64)
	case []float32:
		return v.Value().([]float32)
	case []float64:
		return v.Value().([]float64)
	case []string:
		return v.Value().([]string)
	case []time.Time:
		return v.Value().([]time.Time)
	default:
		log.Errorf("[OPCUA] Unknown value type:%v", v)
		return nil
	}
}

// UaStatusCode2Quality 状态码映射
func UaStatusCode2Quality(code ua.StatusCode) consts.Quality {
	switch code {
	case ua.StatusOK, ua.StatusGood:
		return consts.QualityOk
	default:
		return consts.QualityUncertain
	}
}

// Close 关闭驱动
func (d *Device) Close() consts.Quality {
	log.Warnf("[OPCUA] Close: subs=%d, chans=%d", len(d.subs), len(d.notifyChs))
	for _, sub := range d.subs {
		if sub != nil && d.ctx != nil {
			_ = sub.Cancel(d.ctx)
		}
	}
	for _, ch := range d.notifyChs {
		if ch != nil {
			close(ch)
		}
	}
	d.subs = nil
	d.notifyChs = nil

	if d.client != nil && d.ctx != nil {
		_ = d.client.Close(d.ctx)
		d.client = nil
	}
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}

	// 关闭时丢弃 cache，避免下次误用旧映射
	d.cache = nil

	log.Warnf("[OPCUA] Close done")
	return consts.QualityOk
}

// readNodeIDFromOPCUA 批量读取nodeID的值
func (d *Device) readNodeIDFromOPCUA(ctx context.Context, nodeIDs []string, points []*model.PointInfo) error {
	var nodesToRead []*ua.ReadValueID
	for _, nodeID := range nodeIDs {
		node, err := ua.ParseNodeID(nodeID)
		if err != nil {
			log.Errorf("[OPCUA] parse nodeID fail: %v", err)
			return err
		}
		nodesToRead = append(nodesToRead, &ua.ReadValueID{
			NodeID:      node,
			AttributeID: ua.AttributeIDValue,
		})
	}

	req := &ua.ReadRequest{
		NodesToRead:        nodesToRead,
		MaxAge:             0,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}
	resp, err := d.client.Read(ctx, req)
	if err != nil {
		log.Errorf("[OPCUA] read nodeID fail: %v", err)
		return err
	}

	for i, result := range resp.Results {
		if result.Status != ua.StatusOK {
			points[i].RtVal.Qua = UaStatusCode2Quality(result.Status)
		} else {
			points[i].RtVal.Qua = consts.QualityOk
			points[i].RtVal.Pv = osal.NewVariantWithValue(VariantToGoValue(result.Value))
			tms := result.SourceTimestamp.Unix()
			if result.SourceTimestamp.IsZero() {
				tms = utils.GetNowUTCTimeStamp()
			}
			points[i].RtVal.Tms = tms
		}
	}
	filterLog.Errorf("read"+"done", "[OPCUA] direct read done: nodes=%d", len(nodesToRead))
	return nil
}

// Request 采集流程：cache 命中直接用（刷新上报时间），miss 直读；本 device 的 cache 超时 → 返回触发上层重开
func (d *Device) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model3.MessageStatistics) {
	var msgStat model3.MessageStatistics
	if packet == nil {
		return consts.QualityOk, msgStat
	}
	if d == nil || d.client == nil || d.cache == nil {
		return consts.QualityCommDisconnected, msgStat
	}
	msgStat.SendCount = 1

	if enableSubMod {
		if d.cache.IsTimeout() {
			last := d.cache.GetLastOkTime()
			age := time.Since(last)
			log.Errorf("[OPCUA] cache timeout (this device): last_ok=%v age=%v -> return QualityValueTmsExpired", last, age)
			return consts.QualityValueTmsExpired, msgStat
		}

		var nodeIDs []string
		var points []*model.PointInfo
		var missCount int
		for i := range packet.Points {
			p := packet.Points[i]
			val, ok := d.cache.GetPointValue(string(p.Attr.ID))
			if !ok {
				nodeID, ok := d.cache.GetNodeIDFromReportID(string(p.Attr.ID))
				if !ok {
					log.Errorf("[OPCUA] report id %v not exist", p.Attr.ID)
					p.RtVal.Qua = consts.QualityConfigError
				} else {
					nodeIDs = append(nodeIDs, nodeID)
					points = append(points, p)
					missCount++
				}
			} else {
				p.RtVal = val
				p.RtVal.Tms = utils.GetNowUTCTimeStamp()
			}
		}
		if missCount > 0 {
			filterLog.Errorf("read"+"miss", "[OPCUA] cache miss count=%d, fallback to direct read", missCount)
		}
		if len(nodeIDs) > 0 {
			if err := d.readNodeIDFromOPCUA(ctx, nodeIDs, points); err != nil {
				return consts.QualityCmdRespError, msgStat
			}
		}
	} else {
		// 直采模式：全部走直读
		var nodeIDs []string
		var points []*model.PointInfo
		for i := range packet.Points {
			p := packet.Points[i]
			nodeID, ok := d.cache.GetNodeIDFromReportID(string(p.Attr.ID))
			if !ok {
				log.Errorf("[OPCUA] report id %v not exist", p.Attr.ID)
				p.RtVal.Qua = consts.QualityConfigError
			} else {
				nodeIDs = append(nodeIDs, nodeID)
				points = append(points, p)
			}
		}
		if len(nodeIDs) > 0 {
			if err := d.readNodeIDFromOPCUA(ctx, nodeIDs, points); err != nil {
				return consts.QualityCmdRespError, msgStat
			}
		}
	}

	msgStat.SuccessCount = 1
	return consts.QualityOk, msgStat
}

// RequestPing 请求ping
func (d *Device) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control 控制
func (d *Device) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	return consts.QualityOk
}
