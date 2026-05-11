package kafka

import (
	"agent/entity/config"
	"agent/entity/model/data"
	utils2 "agent/logic/distribution/distributor/utils"
	monitor2 "agent/repo/monitor"
	"agent/utils"
	"agent/utils/thttp"
	"fmt"
	"sync"
	"sync/atomic"

	"trpc.group/trpc-go/trpc-go/log"
)

type retryTarget struct {
	Enable bool
	Name   string
	URL    string
}

type messageRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type messageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func retry(data *data.DataUnit, messages []utils.KafkaMessage,
	kData *utils2.KafkaData, kafkaDataList []*utils2.KafkaData,
	interval int, shouldLog bool, isDefault bool, mozuID string, dataType int) {
	var wg sync.WaitGroup
	successNum := int32(0)

	targets := []retryTarget{
		{
			Enable: config.GetRB().IsBackupPushEnable(),
			Name:   "BackupPush",
			URL:    "", //config.Get().URLPushBackup(),
		},
	}

	for i := range targets {
		if targets[i].Enable {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			t := &targets[i]
			if err := retryBackup(data, messages, kData, kafkaDataList, interval, shouldLog, t, mozuID, dataType); err == nil {
				atomic.AddInt32(&successNum, 1)
			}
		}(i)
	}

	wg.Wait()

	if !isDefault {
		return
	}

	// 至少一次重传成功
	if atomic.LoadInt32(&successNum) > 0 {
		return
	}

	allNames := make([]string, 0, len(targets))
	for i := range targets {
		allNames = append(allNames, targets[i].Name)
	}

	setErrorStatus(data, fmt.Errorf("retry %v 均失败", allNames))
}

func retryBackup(data *data.DataUnit, messages []utils.KafkaMessage,
	kData *utils2.KafkaData, kafkaDataList []*utils2.KafkaData,
	interval int, shouldLog bool, t *retryTarget, mozuID string, dataType int) error {
	if t == nil {
		return nil
	}

	var err error
	for i, msg := range messages {
		err = DistributeToBackup(t.Name, t.URL, string(msg.Key), string(msg.Value))
		if err != nil && shouldLog {
			monitor2.LogReportError(err)
		}
		if err != nil {
			break
		}
		if shouldLog {
			kafkaDataList[i].Log(t.Name, data.DeviceGid, interval)
		}
	}
	if err == nil {
		kData.Report(fmt.Sprintf("http-%v", t.Name), interval, mozuID, dataType)
		return nil
	}

	log.Errorf("Distribute to %v error: %v", t.Name, err)
	return err
}

// DistributeToBackup 分发到备份
func DistributeToBackup(name string, url string, key string, value string) error {
	req := messageRequest{
		Key:   key,
		Value: value,
	}
	return thttp.PostJSON(url, req, 60000, nil)
}
