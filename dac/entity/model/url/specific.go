// Package url 提供门禁控制器HTTP接口的URL生成器。
package url

import (
	"fmt"
)

// SpecificProducer 自研门禁控制器的URL生成器
type SpecificProducer struct {
	baseInfo BaseInfo
}

// NewSpecificProducer 创建自研门禁URL生成器实例
func NewSpecificProducer(
	channelID string, apiKey string,
) *SpecificProducer {
	return &SpecificProducer{baseInfo: BaseInfo{
		ChannelID: channelID,
		ApiKey:    apiKey,
	}}
}

// GetDoorPositionStateURL 获取门位置状态接口URL
func (d *SpecificProducer) GetDoorPositionStateURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_doorposition_state&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetDoorStateURL 获取门状态接口URL
func (d *SpecificProducer) GetDoorStateURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_door_state&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// SetDoorStateURL 设置门状态接口URL
func (d *SpecificProducer) SetDoorStateURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_door_state&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetDoorsURL 获取门列表接口URL
func (d *SpecificProducer) GetDoorsURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_doors&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetDoorParameterURL 获取门参数接口URL
func (d *SpecificProducer) GetDoorParameterURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_doorpara&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// SetDoorParameterURL 设置门参数接口URL
func (d *SpecificProducer) SetDoorParameterURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_doorpara&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetHistoryEventHisURL 获取历史事件接口URL（从最早时间开始）
func (d *SpecificProducer) GetHistoryEventHisURL(
	recordIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=2000-01-01%%2008:00:00&time_end=-1"+
			"&record_index=%v&alarm_index=0&apikey=%v",
		d.baseInfo.ChannelID, recordIndex, d.baseInfo.ApiKey)
}

// GetHistoryEventByTimestampURL 按时间范围获取历史事件接口URL
func (d *SpecificProducer) GetHistoryEventByTimestampURL(
	begin, end string, recordIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=%v&time_end=%v"+
			"&record_index=%v&alarm_index=0&apikey=%v",
		d.baseInfo.ChannelID, begin, end,
		recordIndex, d.baseInfo.ApiKey)
}

// GetMDCEventURL 获取MDC事件接口URL
func (d *SpecificProducer) GetMDCEventURL(
	recordIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_event"+
			"&record_index=%v&alarm_index=0&apikey=%v",
		d.baseInfo.ChannelID, recordIndex, d.baseInfo.ApiKey)
}

// GetHistoryAlarmURL 获取历史告警接口URL（从最早时间开始）
func (d *SpecificProducer) GetHistoryAlarmURL(
	alarmIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=2000-01-01%%2008:00:00&time_end=-1"+
			"&record_index=0&alarm_index=%v&apikey=%v",
		d.baseInfo.ChannelID, alarmIndex, d.baseInfo.ApiKey)
}

// GetHistoryAlarmByTimestampURL 按时间范围获取历史告警接口URL
func (d *SpecificProducer) GetHistoryAlarmByTimestampURL(
	begin, end string, alarmIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_history_event_by_batch"+
			"&time_begin=%v&time_end=%v"+
			"&record_index=0&alarm_index=%v&apikey=%v",
		d.baseInfo.ChannelID, begin, end,
		alarmIndex, d.baseInfo.ApiKey)
}

// GetMDCAlarmURL 获取MDC告警接口URL
func (d *SpecificProducer) GetMDCAlarmURL(
	alarmIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_event"+
			"&record_index=0&alarm_index=%v&apikey=%v",
		d.baseInfo.ChannelID, alarmIndex, d.baseInfo.ApiKey)
}

// GetTimeGroupURL 获取时间组接口URL
func (d *SpecificProducer) GetTimeGroupURL(
	groupNo interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_timegroup"+
			"&group_no=%v&apikey=%v",
		d.baseInfo.ChannelID, groupNo, d.baseInfo.ApiKey)
}

// SetTimeGroupURL 设置时间组接口URL
func (d *SpecificProducer) SetTimeGroupURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_timegroup&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// ClearTimeGroupURL 清除时间组接口URL
func (d *SpecificProducer) ClearTimeGroupURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=clear_timegroup&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetTimeURL 获取当前时间接口URL
func (d *SpecificProducer) GetTimeURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_current_time&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// SetTimeURL 设置当前时间接口URL
func (d *SpecificProducer) SetTimeURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=set_current_time&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetCardsURL 分页获取卡列表接口URL
func (d *SpecificProducer) GetCardsURL(
	cardIndex interface{},
) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_cards_by_batch"+
			"&card_index=%v&apikey=%v",
		d.baseInfo.ChannelID, cardIndex, d.baseInfo.ApiKey)
}

// GetAllCardsURL 获取所有卡接口URL
func (d *SpecificProducer) GetAllCardsURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_cards&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// AddCardURL 添加卡接口URL
func (d *SpecificProducer) AddCardURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=add_card&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetCardURL 获取指定卡号的卡信息接口URL
func (d *SpecificProducer) GetCardURL(cardNo string) string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_card&card_no=%v&apikey=%v",
		d.baseInfo.ChannelID, cardNo, d.baseInfo.ApiKey)
}

// UpdateCardURL 更新卡接口URL
func (d *SpecificProducer) UpdateCardURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=modify_card&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// DeleteCardURL 删除卡接口URL
func (d *SpecificProducer) DeleteCardURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=delete_card&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// CleanURL 清空设备数据接口URL
func (d *SpecificProducer) CleanURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=clean_device&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// AddUserURL 添加用户接口URL（V3协议）
func (d *SpecificProducer) AddUserURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=add_user&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// DeleteUserURL 删除用户接口URL（V3协议）
func (d *SpecificProducer) DeleteUserURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=delete_user&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// ResetURL 重置设备接口URL
func (d *SpecificProducer) ResetURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=reset_device&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// GetCurrentAlarmURL 获取当前告警接口URL
func (d *SpecificProducer) GetCurrentAlarmURL() string {
	return fmt.Sprintf(
		"http://%v/dac?action=get_door_alarm&apikey=%v",
		d.baseInfo.ChannelID, d.baseInfo.ApiKey)
}

// NeedDataPrefix 是否需要数据前缀（自研协议不需要）
func (d *SpecificProducer) NeedDataPrefix() bool {
	return false
}
