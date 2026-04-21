// Package url 提供门禁控制器HTTP接口的URL生成器。
package url

import (
	"fmt"
)

// DefaultProducer 默认门禁控制器的URL生成器
type DefaultProducer struct {
	baseInfo BaseInfo
}

// NewDefaultProducer 创建默认门禁URL生成器实例
func NewDefaultProducer(
	channelID string, apiKey string,
) *DefaultProducer {
	return &DefaultProducer{baseInfo: BaseInfo{
		ChannelID: channelID,
		ApiKey:    apiKey,
	}}
}

// GetDoorPositionStateURL 获取门位置状态接口URL
func (d *DefaultProducer) GetDoorPositionStateURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_doorposition_state&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetDoorStateURL 获取门状态接口URL
func (d *DefaultProducer) GetDoorStateURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_door_state&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// SetDoorStateURL 设置门状态接口URL
func (d *DefaultProducer) SetDoorStateURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_door_state&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetDoorsURL 获取门列表接口URL
func (d *DefaultProducer) GetDoorsURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_doors&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetDoorParameterURL 获取门参数接口URL
func (d *DefaultProducer) GetDoorParameterURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_doorpara&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// SetDoorParameterURL 设置门参数接口URL
func (d *DefaultProducer) SetDoorParameterURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_doorpara&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetHistoryEventHisURL 获取历史事件接口URL（从最早时间开始）
func (d *DefaultProducer) GetHistoryEventHisURL(
	recordIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=2000-01-01%%2008:00:00&time_end=-1"+
			"&apikey=%v&record_index=%v&alarm_index=0",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, recordIndex)
}

// GetHistoryEventByTimestampURL 按时间范围获取历史事件接口URL
func (d *DefaultProducer) GetHistoryEventByTimestampURL(
	begin, end string, recordIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=%v&time_end=%v"+
			"&apikey=%v&record_index=%v&alarm_index=0",
		d.baseInfo.ChannelID, begin, end,
		d.baseInfo.ApiKey, recordIndex)
}

// GetMDCEventURL 获取MDC事件接口URL
func (d *DefaultProducer) GetMDCEventURL(
	recordIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_event"+
			"&apikey=%v&record_index=%v&alarm_index=0",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, recordIndex)
}

// GetHistoryAlarmURL 获取历史告警接口URL（从最早时间开始）
func (d *DefaultProducer) GetHistoryAlarmURL(
	alarmIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=2000-01-01%%2008:00:00&time_end=-1"+
			"&apikey=%v&record_index=0&alarm_index=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, alarmIndex)
}

// GetHistoryAlarmByTimestampURL 按时间范围获取历史告警接口URL
func (d *DefaultProducer) GetHistoryAlarmByTimestampURL(
	begin, end string, alarmIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=%v&time_end=%v"+
			"&apikey=%v&record_index=0&alarm_index=%v",
		d.baseInfo.ChannelID, begin, end,
		d.baseInfo.ApiKey, alarmIndex)
}

// GetMDCAlarmURL 获取MDC告警接口URL
func (d *DefaultProducer) GetMDCAlarmURL(
	alarmIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_event"+
			"&apikey=%v&record_index=0&alarm_index=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, alarmIndex)
}

// GetTimeGroupURL 获取时间组接口URL
func (d *DefaultProducer) GetTimeGroupURL(
	groupNo interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_timegroup"+
			"&apikey=%v&group_no=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, groupNo)
}

// SetTimeGroupURL 设置时间组接口URL
func (d *DefaultProducer) SetTimeGroupURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_timegroup&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// ClearTimeGroupURL 清除时间组接口URL
func (d *DefaultProducer) ClearTimeGroupURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=clear_timegroup&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetTimeURL 获取当前时间接口URL
func (d *DefaultProducer) GetTimeURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_current_time&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// SetTimeURL 设置当前时间接口URL
func (d *DefaultProducer) SetTimeURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_current_time&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetCardsURL 分页获取卡列表接口URL
func (d *DefaultProducer) GetCardsURL(
	cardIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_cards_by_batch"+
			"&apikey=%v&card_index=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, cardIndex)
}

// GetAllCardsURL 获取所有卡接口URL
func (d *DefaultProducer) GetAllCardsURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_cards&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// AddCardURL 添加卡接口URL
func (d *DefaultProducer) AddCardURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=add_card&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetCardURL 获取指定卡号的卡信息接口URL
func (d *DefaultProducer) GetCardURL(cardNo string) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_card&apikey=%v&card_no=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey, cardNo)
}

// UpdateCardURL 更新卡接口URL
func (d *DefaultProducer) UpdateCardURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=modify_card&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// DeleteCardURL 删除卡接口URL
func (d *DefaultProducer) DeleteCardURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=delete_card&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// AddUserURL 添加用户接口URL（V3协议）
func (d *DefaultProducer) AddUserURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=add_user&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// DeleteUserURL 删除用户接口URL（V3协议）
func (d *DefaultProducer) DeleteUserURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=delete_user&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// CleanURL 清空设备数据接口URL
func (d *DefaultProducer) CleanURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=clean_device&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// ResetURL 重置设备接口URL
func (d *DefaultProducer) ResetURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=reset_device&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetCurrentAlarmURL 获取当前告警接口URL
func (d *DefaultProducer) GetCurrentAlarmURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_door_alarm&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// NeedDataPrefix 是否需要数据前缀（默认协议需要）
func (d *DefaultProducer) NeedDataPrefix() bool {
	return true
}
