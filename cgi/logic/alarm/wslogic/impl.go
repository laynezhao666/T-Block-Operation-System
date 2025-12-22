// Package wslogic wslogic
package wslogic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go"

	"etrpc-go/log"

	"github.com/gorilla/websocket"
	"github.com/samber/lo"

	"cgi/conf"
	"cgi/entity/model"
	"cgi/entity/wsmodel"
	"cgi/logic/alarm/api"

	pb "trpcprotocol/cgi"
)

var (
	once        sync.Once
	alarmWSImpl *AlarmWSImpl
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

const (
	// WS接口告警推送 Cmd
	alarmPushCmd = "getInitDetailList"
)

// GetAlarmWSImpl GetAlarmWSImpl
func GetAlarmWSImpl() *AlarmWSImpl {
	once.Do(func() {
		impl := &AlarmWSImpl{
			mozuMap: make(map[int32][]*websocket.Conn),
			connMap: make(map[string]int32),
			taskCh:  make(chan *wsmodel.DataMozu, conf.AlarmConfImpl.WSTaskChannelSize),
		}
		alarmWSImpl = impl
	})
	return alarmWSImpl
}

// AlarmWSImpl 需要实现 (trpc-tnet-transport/websocket).Service interface.
type AlarmWSImpl struct {
	mu      sync.RWMutex
	mozuMap map[int32][]*websocket.Conn // mozuId connList
	connMap map[string]int32            // conn mozuId
	taskCh  chan *wsmodel.DataMozu
}

type infoKey struct{}

type connectInfo struct {
	mozuId int32
}

// RecvMsg RecvMsg
type RecvMsg struct {
	Reqid     int    `json:"reqid"`
	Cmd       string `json:"cmd"`
	Timestamp int64  `json:"timestamp"`
}

func (a *AlarmWSImpl) addConn(conn *websocket.Conn, mozuId int32) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.mozuMap[mozuId]; !ok {
		a.mozuMap[mozuId] = make([]*websocket.Conn, 0)
	}
	a.mozuMap[mozuId] = append(a.mozuMap[mozuId], conn)
	a.connMap[conn.RemoteAddr().String()] = mozuId
	return nil
}

func (a *AlarmWSImpl) stopConn(conn *websocket.Conn) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	addr := conn.RemoteAddr().String()
	var mozuId int32
	if _, ok := a.connMap[addr]; !ok {
		return nil
	} else {
		mozuId = a.connMap[addr]
	}
	delete(a.connMap, addr)
	if _, ok := a.mozuMap[mozuId]; !ok {
		return nil
	}
	for i, c := range a.mozuMap[mozuId] {
		if c.RemoteAddr().String() == addr {
			a.mozuMap[mozuId] = append(a.mozuMap[mozuId][0:i], a.mozuMap[mozuId][i+1:]...)
			log.Infof("After Stop, mozuId:%v, curLen:%d", mozuId, len(a.mozuMap[mozuId]))
			if len(a.mozuMap[mozuId]) == 0 {
				delete(a.mozuMap, mozuId)
			}
			break
		}
	}
	return nil
}

// HandleWebSocket 处理websocket请求
func (a *AlarmWSImpl) HandleWebSocket(w http.ResponseWriter, r *http.Request) error {
	protocols := websocket.Subprotocols(r)
	if len(protocols) == 0 {
		http.Error(w, "mozuid not found in Sec-WebSocket-Protocol", http.StatusBadRequest)
		return fmt.Errorf("mozuid not found")
	}
	mozuidStr := protocols[0]
	mozuId, err := strconv.Atoi(mozuidStr)
	if err != nil || mozuId <= 0 {
		mozuId, err = strconv.Atoi(r.Header.Get("mozu_id"))
		if err != nil || mozuId <= 0 {
			log.Errorf("parse mozuId failed, errMsg:%v", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return fmt.Errorf("Forbidden")
		}
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade failed:", err)
		return fmt.Errorf("WebSocket upgrade failed: %v", err)
	}
	defer conn.Close()
	// 设置心跳机制
	conn.SetPingHandler(func(message string) error {
		return conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(3*time.Second))
	})
	if err := a.addConn(conn, int32(mozuId)); err != nil {
		log.Errorf("addConn failed, errMsg:%v", err)
		return fmt.Errorf("addConn failed: %v", err)
	}
	log.Infof("Connected conn:%v, mozuId:%v, curLen:%d",
		conn.RemoteAddr().String(), mozuId, len(a.mozuMap[int32(mozuId)]))
	// 处理消息
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("WebSocket read failed:", err)
			conn.Close()
			break
		}
		recvData := &RecvMsg{}
		if err := json.Unmarshal(data, recvData); err != nil {
			log.Warnf("Read failed, msg:%s, err:%v", string(data), err)
			continue
		}
		if recvData.Cmd == alarmPushCmd {
			a.pushToConn(trpc.BackgroundContext(), conn)
		}
	}
	log.Infof("Disconnected conn:%v, mozuId:%v, curLen:%d",
		conn.RemoteAddr().String(), mozuId, len(a.mozuMap[int32(mozuId)]))
	if err := a.stopConn(conn); err != nil {
		log.Errorf("stopConn failed, errMsg:%v", err)
	}
	return nil
}

func getPushData(ctx context.Context, mozuId int32) (*wsmodel.ResponseData, error) {
	req := &pb.ReqAlarmList{
		MozuId:    int32(mozuId),
		AlarmType: 1,
	}
	rsp, err := api.NewAlarmApi().GetAlarmList(ctx, req)
	if err != nil {
		log.Errorf("GetAlarmList failed to getPushData, errMsg:%v", err)
		return nil, err
	}
	if len(rsp.List) == 0 {
		return nil, nil
	}
	data := make(map[string]interface{})
	data["warn"] = lo.Map(rsp.List, func(item *pb.RspAlarmList_Item, index int) *model.AlarmActiveWS {
		return &model.AlarmActiveWS{
			AlarmId:      strconv.FormatInt(item.AlarmId, 10),
			Level:        item.Level,
			AlarmName:    item.AlarmName,
			Rid:          item.Rid,
			DeviceGid:    item.DeviceGid,
			DeviceNumber: item.DeviceNumber,
			DeviceTypeZh: item.DeviceTypeZh,
			Box:          item.Box,
			Room:         item.Room,
			MozuId:       item.MozuId,
			MozuName:     item.MozuName,
			AlarmContent: item.AlarmContent,
			AlarmStatus:  item.AlarmStatus,
			EventStatus:  item.EventStatus,
			Points:       item.Points,
			OccurTime:    item.OccurTime,
		}
	})
	wsRsp := &wsmodel.ResponseData{
		Code:      0,
		Data:      data,
		Cmd:       alarmPushCmd,
		Timestamp: time.Now().Unix(),
	}
	return wsRsp, nil
}

// PushToConn 用来推送下行消息
func (s *AlarmWSImpl) pushToConn(ctx context.Context, conn *websocket.Conn) error {
	addr := conn.RemoteAddr().String()
	if _, ok := s.connMap[addr]; !ok {
		return nil
	}
	mozuId := s.connMap[addr]
	wsRsp, err := getPushData(ctx, mozuId)
	if err != nil {
		log.Errorf("getPushData failed PushToConn, errMsg:%v", err)
		return err
	}
	if wsRsp == nil {
		return nil
	}
	wsRspByte, err := json.Marshal(wsRsp)
	if err != nil {
		log.Errorf("Marshal failed PushToConn, errMsg:%v", err)
		return err
	}
	if err := conn.WriteMessage(websocket.TextMessage, wsRspByte); err != nil {
		log.Errorf("WriteMessage failed PushToConn, errMsg:%v", err)
		return err
	}
	return nil
}

// ExecPushAlarm 通道推送告警信息
func (s *AlarmWSImpl) ExecPushAlarm(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return nil
		case data := <-s.taskCh:
			mozuId := data.MozuId
			s.mu.RLock()
			connList, ok := s.mozuMap[mozuId]
			if !ok {
				log.Warnf("No socket listening, mozuId:%d", mozuId)
			}
			for _, conn := range connList {
				if err := conn.WriteMessage(websocket.TextMessage, data.Data); err != nil {
					log.Warnf("WriteMessage failed ExecPushAlarm, addr: %s, errMsg:%v", conn.RemoteAddr().String(), err)
					continue
				}
			}
			s.mu.RUnlock()
		}
	}
}

// AddMozuPushTask 添加模组推送任务
func (s *AlarmWSImpl) AddMozuPushTask(mozuId int32) error {
	wsRsp, err := getPushData(trpc.BackgroundContext(), mozuId)
	if err != nil {
		log.Errorf("getPushData failed AddMozuPushTask, mozuId: %d, errMsg:%v", mozuId, err)
		return err
	}
	if wsRsp == nil {
		return nil
	}
	wsRspByte, err := json.Marshal(wsRsp)
	if err != nil {
		log.Errorf("Marshal failed AddMozuPushTask, mozuId: %d, errMsg:%v", mozuId, err)
		return err
	}
	pushData := &wsmodel.DataMozu{
		MozuId: mozuId,
		Data:   wsRspByte,
	}
	s.taskCh <- pushData
	return nil
}

// RegularPushAll 定时推送所有模组
func (s *AlarmWSImpl) RegularPushAll(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()
	interval := conf.AlarmConfImpl.PushWSInterval
	if interval == 0 {
		interval = 10
	}
	itv := time.Second * time.Duration(interval)
	tick := time.NewTicker(itv)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			s.mu.RLock()
			mozuIdList := []int32{}
			for mozuId := range s.mozuMap {
				mozuIdList = append(mozuIdList, mozuId)
			}
			s.mu.RUnlock()
			for _, mozuId := range mozuIdList {
				s.AddMozuPushTask(mozuId)
			}
		}
	}
}
