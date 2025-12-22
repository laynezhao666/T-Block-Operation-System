package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"etrpc-go/log"

	"github.com/avast/retry-go"
	"github.com/panjf2000/ants/v2"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/conf"
	"alarm-manage/entity/message"
	"alarm-manage/logic/notification"
	"alarm-manage/logic/snowflake"
	"alarm-manage/repo/cache"
	"alarm-manage/repo/db"
	"alarm-manage/utils/batch"
	cmodel "common/entity/model"
)

var (
	AlarmContentRegex *regexp.Regexp = regexp.MustCompile("{{(.*?)}}")
)

func (m *Manager) processFireAlert(ctx context.Context) {
	var doJobTask = func(i interface{}) error {
		activeList := i.([]*cmodel.AlarmActive)
		if len(activeList) == 0 {
			return nil
		}
		m.processEdgeFireAlert(ctx, activeList)
		return nil
	}
	var poolWg sync.WaitGroup
	wpSize := conf.ServerConf.AlertManageConfig.PoolSize
	wp, _ := ants.NewPoolWithFunc(int(wpSize), func(i interface{}) {
		doJobTask(i)
		poolWg.Done()
	}, ants.WithNonblocking(false))
	defer wp.Release()
	size := conf.ServerConf.AlertManageConfig.BatchChannelSize
	interval := time.Duration(conf.ServerConf.AlertManageConfig.BatchFetchIntervalMS) * time.Millisecond
	activeChannel := batch.BatchChannel(ctx, m.alertingCh, int(size), interval)
	for {
		select {
		case <-ctx.Done():
			poolWg.Wait()
			return
		case activeList := <-activeChannel:
			edgeAlarms := m.splitActiveAlarms(activeList)
			poolWg.Add(1)
			wp.Invoke(edgeAlarms)
		}
	}
}

func (m *Manager) processEdgeFireAlert(ctx context.Context, actives []*cmodel.AlarmActive) (err error) {
	dbActive, err := BatchAddActiveAlert(actives)
	log.Infof("receive active alerts, actives len: %d, dbActive len: %d", len(actives), len(dbActive))
	if err != nil {
		// 如果写失败了，需要将告警打印出来，以便需要手动补告警
		bjson, _ := json.Marshal(actives)
		log.AlarmContextf(trpc.BackgroundContext(), "活动告警插入数据库失败, alert: %v, error: %s", string(bjson), err.Error())
	}
	// TODO 告警后续处理，上报活动告警等
	if len(dbActive) > 0 {
		go notification.GetSpeaker().ReportAlert(dbActive)
	}
	return nil
}

func (m *Manager) splitActiveAlarms(list *batch.GroupData) (edgeAlarms []*cmodel.AlarmActive) {
	for _, item := range list.Data {
		ac := item.(*cmodel.AlarmActive)
		if !isAlarmActiveValid(ac) {
			continue
		}
		edgeAlarms = append(edgeAlarms, ac)
	}
	return
}

func isAlarmActiveValid(ac *cmodel.AlarmActive) bool {
	if ac.DeviceGid == "" || ac.Rid == 0 {
		return false
	}
	return true
}

// BatchAddActiveAlert BatchAddActiveAlert
func BatchAddActiveAlert(actives []*cmodel.AlarmActive) ([]cmodel.AlarmActive, error) {
	var err error
	dbList := make([]cmodel.AlarmActive, 0)
	// 过滤指纹，已存在指纹不插入
	insertList, err := getUniqueActiveList(actives)
	if err != nil {
		// retry
		log.Warnf("getUniqueActiveList failed, retry, err: %v", err)
		time.Sleep(100 * time.Microsecond)
		insertList, err = getUniqueActiveList(actives)
		if err != nil {
			return dbList, err
		}
	}
	if len(insertList) == 0 {
		// filterActiveByFingerprint 方法已打印了活动告警中已有的 fp 告警，不再重复打印
		log.Infof("all actives existed")
		return dbList, nil
	}
	// 填充字段
	err = fillActivesAlert(insertList)
	if err != nil {
		// retry
		log.Warnf("fillActivesAlert failed, retry, err: %v", err)
		time.Sleep(100 * time.Microsecond)
		err = fillActivesAlert(insertList)
		if err != nil {
			return dbList, err
		}
	}
	dbList, err = retryAddActiveAlertMutil(insertList)
	if err != nil {
		return dbList, err
	}
	return dbList, nil
}

func getUniqueActiveList(actives []*cmodel.AlarmActive) ([]*cmodel.AlarmActive, error) {
	list := make(map[string]*cmodel.AlarmActive)
	// 添加指纹
	for _, item := range actives {
		if v, ok := list[item.FingerPrint]; ok {
			log.Infof("fp duplicated, v: %v, item: %v", *v, *item)
			if v.UpdateTime.Before(item.UpdateTime) {
				v.UpdateTime = item.UpdateTime
			}
			if v.OccurTime.After(item.OccurTime) {
				v.OccurTime = item.OccurTime
			}
			list[item.FingerPrint] = v
		} else {
			// 不存在，保存起来
			list[item.FingerPrint] = item
		}
	}
	// 去重本批
	uniqueFpList := make([]string, 0)
	uniqueList := make([]*cmodel.AlarmActive, 0)
	for _, item := range list {
		uniqueList = append(uniqueList, item)
		uniqueFpList = append(uniqueFpList, item.FingerPrint)
	}
	log.Infof("allList len: %d, unique len:%d", len(list), len(uniqueList))
	// 过滤指纹，已存在指纹不插入
	return filterActiveByFingerprint(uniqueFpList, uniqueList)
}

func filterActiveByFingerprint(fps []string, list []*cmodel.AlarmActive) ([]*cmodel.AlarmActive, error) {
	var err error
	_, notExistFp, err := checkActiveExistByFingerprints(fps)
	if err != nil {
		return nil, err
	}
	exist := make([]*cmodel.AlarmActive, 0)
	notExist := make([]*cmodel.AlarmActive, 0)
	notExistlistMap := make(map[string]struct{})
	for _, item := range notExistFp {
		notExistlistMap[item] = struct{}{}
	}
	for _, item := range list {
		if _, ok := notExistlistMap[item.FingerPrint]; ok {
			notExist = append(notExist, item)
		} else {
			exist = append(exist, item)
		}
	}
	return notExist, nil
}

func checkActiveExistByFingerprints(fingerprints []string) (exist, notExist []string, err error) {

	exist = make([]string, 0)
	notExist = make([]string, 0)
	// 只获取 fingerprint，减少数据传输
	list, err := db.GetAlarmDBImpl().GetActiveFingerprints(fingerprints)
	if err != nil {
		return exist, notExist, err
	}
	fmap := make(map[string]struct{})
	for _, item := range list {
		fmap[item] = struct{}{}
	}
	for _, item := range fingerprints {
		if _, ok := fmap[item]; ok {
			exist = append(exist, item)
		} else {
			notExist = append(notExist, item)
		}
	}
	return exist, notExist, nil
}

func fillActivesAlert(list []*cmodel.AlarmActive) error {
	var err error
	// 填充创建时间
	t := time.Now()
	lo.ForEach(list, func(item *cmodel.AlarmActive, index int) {
		item.CreateAt = t
	})
	// 检查设备是否存在,填充设备类型字段
	err = filterActiveListByDeviceInfo(list)
	if err != nil {
		return err
	}
	batchFillContentWithVar(list)
	return nil
}

// TODO 填充设备信息： 方舱名、房间号、设备类型......
func filterActiveListByDeviceInfo(list []*cmodel.AlarmActive) error {
	for _, item := range list {
		gid := item.DeviceGid
		if len(gid) == 0 {
			continue
		}
		entity, ok := cache.GetLocalCache().GetDeviceCache(gid)
		if !ok {
			continue
		}
		item.BoxName = entity.FuncRoom
		item.RoomName = entity.IdcArea
		item.DeviceNumber = entity.DeviceNumber
		item.DeviceName = entity.DeviceName
		item.DeviceTypeZh = entity.DeviceTypeZh
		item.MozuName = entity.MozuName
		item.DeviceTypeEn = entity.DeviceTypeEn
	}
	return nil
}

func getAlarmRet(retStr string) (ret message.AlarmTaskRet, err error) {
	ret = message.AlarmTaskRet{}
	if err = json.Unmarshal([]byte(retStr), &ret); err != nil {
		return
	}

	return
}

// batchFillContentVar 填充告警内容中的测点变量 {{A}}
//
// * 将变量映射为测点
// * 获取测点的数值
// * 将测点替换为数值
func batchFillContentWithVar(activeList []*cmodel.AlarmActive) {
	for _, item := range activeList {
		content := item.Content
		if len(content) == 0 {
			continue
		}
		// 正则表达式取出标准测点
		matches := AlarmContentRegex.FindAllStringSubmatch(content, -1)
		ret, err := getAlarmRet(item.AnalyzeResult)
		if err != nil {
			log.Warnf("fill content failed, alarmId: %v", item.AlarmID)
			continue
		}
		pMap := ret.PointMap
		pointList := [][]string{}
		for _, matchItem := range matches {
			if pointNames, ok := pMap[matchItem[1]]; ok {
				pointList = append(pointList, pointNames)
			}
		}
		content, err = replaceContent(content, matches, pointList, ret)
		if err != nil {
			log.Warnf("fill content failed, alarmId: %v", item.AlarmID)
			continue
		}
		item.Content = content
	}
}

// 若为实时策略，则从PointValueMap取值
// 若为延时策略，从HistoryPointValueMap取值
func replaceContent(content string, matches [][]string, pointsList [][]string, ret message.AlarmTaskRet) (string, error) {
	if ret.PointValueMap != nil {
		for index, points := range pointsList {
			values := []string{}
			for _, p := range points {
				val, ok := ret.PointValueMap[p]
				if !ok {
					return content, fmt.Errorf("point value not found, pointName: %s", p)
				}
				values = append(values, strconv.FormatFloat(val, 'f', 2, 64))
			}
			content = strings.ReplaceAll(content, matches[index][0], strings.Join(values, ","))
		}
	} else {
		for index, points := range pointsList {
			values := []string{}
			for _, p := range points {
				val, ok := ret.HistoryPointValueMap[p][0]
				if !ok {
					return content, fmt.Errorf("point value not found, pointName: %s", p)
				}
				values = append(values, strconv.FormatFloat(val, 'f', 2, 64))
			}
			content = strings.ReplaceAll(content, matches[index][0], strings.Join(values, ","))
		}
	}
	return content, nil
}

// 1. 活动告警写入数据库
// 2. 筛选成功写入的告警
// 3. 返回，用于告警通知
func retryAddActiveAlertMutil(active []*cmodel.AlarmActive) ([]cmodel.AlarmActive, error) {
	var err error
	dbList, err := insertActiveMutil(active)
	if err != nil {
		// 重试逻辑
		bjson, _ := json.Marshal(active)
		log.Errorf("add alert to db error,retry one times, active: %s, err:%s",
			string(bjson), err.Error())
		time.Sleep(100 * time.Microsecond)
		dbList, err = insertActiveMutil(active)
		if err != nil {
			return dbList, err
		}
	}
	return dbList, nil
}

func insertActiveMutil(list []*cmodel.AlarmActive) ([]cmodel.AlarmActive, error) {
	dbList := []cmodel.AlarmActive{}
	for _, item := range list {
		// 雪花算法生成告警ID
		nodeID, err := snowflake.GenerateAlarmId()
		if err != nil {
			log.Errorf("fillAlarmID failed, err: %v, active:%v", err, item)
			continue
		}
		item.AlarmID = nodeID.Int64()
		dbList = append(dbList, *item)
	}
	// 写入失败 重试三次
	var successAlarmList []cmodel.AlarmActive
	err := retry.Do(func() error {
		var retryErr error
		successAlarmList, retryErr = db.GetAlarmDBImpl().BatchInsertActiveAlerts(trpc.BackgroundContext(), dbList)
		return retryErr
	}, retry.Attempts(3), retry.RetryIf(func(retryErr error) bool {
		return retryErr != nil
	}))
	if err != nil {
		return successAlarmList, err
	}
	return successAlarmList, nil
}
