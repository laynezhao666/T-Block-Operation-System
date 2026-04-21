package dac

import (
	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// withTimestampDesc 按时间戳降序排序。
func withTimestampDesc(tx *gorm.DB) *gorm.DB {
	return withDESC(tx, "timestamp")
}

// withTimestampBetween 构造时间戳范围查询条件。
func withTimestampBetween(tx *gorm.DB, begin, end interface{}) *gorm.DB {
	return withBetween(tx, "timestamp", begin, end)
}

// withID 构造 id = value 查询条件。
func withID(tx *gorm.DB, id interface{}) *gorm.DB {
	return withEqual(tx, "id", id)
}

// withIDs 构造 id IN (values) 查询条件。
func withIDs(tx *gorm.DB, ids interface{}) *gorm.DB {
	return withIn(tx, "id", ids)
}

// withAccessGroupID 构造权限组ID等值查询条件。
func withAccessGroupID(tx *gorm.DB, id interface{}) *gorm.DB {
	return withEqual(tx, db.ColumnAccessGroupID, id)
}

// withAccessGroupIDs 构造权限组ID列表查询条件。
func withAccessGroupIDs(tx *gorm.DB, ids interface{}) *gorm.DB {
	return withIn(tx, db.ColumnAccessGroupID, ids)
}

// withDoorIDs 构造门ID列表查询条件。
func withDoorIDs(tx *gorm.DB, ids interface{}) *gorm.DB {
	return withIn(tx, "door_id", ids)
}

// withControllerIDs 构造门禁控制器ID列表查询条件。
func withControllerIDs(tx *gorm.DB, ids interface{}) *gorm.DB {
	return withIn(tx, "controller_id", ids)
}

// queryAndCountRecords 分页查询记录并统计总数。
func queryAndCountRecords(tx *gorm.DB, offset, limit int, values interface{}, count *int64) error {
	return tx.Offset(offset).Limit(limit).Find(values).Offset(-1).Limit(-1).Count(count).Error
}

// withASC 按指定字段升序排序（委托给 tgorm 包，含字段名安全校验）。
func withASC(tx *gorm.DB, field string) *gorm.DB {
	return tgorm.WithASC(field)(tx)
}

// withDESC 按指定字段降序排序（委托给 tgorm 包，含字段名安全校验）。
func withDESC(tx *gorm.DB, field string) *gorm.DB {
	return tgorm.WithDESC(field)(tx)
}

// countRecord 统计符合条件的记录总数。
func countRecord(tx *gorm.DB, count *int64) error {
	return tx.Count(count).Error
}

// withBetween 构造字段范围查询条件（委托给 tgorm 包，含字段名安全校验）。
func withBetween(tx *gorm.DB, filed string, begin, end interface{}) *gorm.DB {
	return tgorm.WithBetween(filed, begin, end)(tx)
}

// withEqual 构造字段等值查询条件（委托给 tgorm 包，含字段名安全校验）。
func withEqual(tx *gorm.DB, field string, value interface{}) *gorm.DB {
	return tgorm.WithEqual(field, value)(tx)
}

// withIn 构造字段 IN 查询条件（委托给 tgorm 包，含字段名安全校验）。
func withIn(tx *gorm.DB, field string, values interface{}) *gorm.DB {
	return tgorm.WithIn(field, values)(tx)
}

// queryRecordByID 根据ID查询单条记录。
func queryRecordByID(tx *gorm.DB, id interface{}, value interface{}) error {
	return withID(tx, id).First(value).Error
}

// queryRecordsByIDs 根据ID列表查询多条记录。
func queryRecordsByIDs(tx *gorm.DB, ids interface{}, values interface{}) error {
	return withIDs(tx, ids).Find(values).Error
}

// deleteRecordsByID 根据ID列表删除记录。
func deleteRecordsByID(tx *gorm.DB, ids interface{}, values interface{}) error {
	return withIDs(tx, ids).Delete(values).Error
}

// withName 构造名称等值查询 Option。
func withName(name string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnName, name)
}

// withIDOption 构造ID等值查询 Option。
func withIDOption(id interface{}) tgorm.Option {
	return tgorm.WithEqual(db.ColumnID, id)
}

// withIDsOption 构造ID列表查询 Option。
func withIDsOption(ids interface{}) tgorm.Option {
	return tgorm.WithIn(db.ColumnID, ids)
}

// withMozuID 构造模组ID等值查询 Option。
func withMozuID(id interface{}) tgorm.Option {
	return tgorm.WithEqual(db.ColumnMozuID, id)
}

// withMethod 构造请求方法等值查询 Option。
func withMethod(method string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnMethod, method)
}

// withBuildingMID 构造楼栋MID等值查询 Option。
func withBuildingMID(mid int) tgorm.Option {
	return tgorm.WithEqual(db.ColumnBuildingMID, mid)
}

// withControllerIDsOption 构造门禁控制器ID列表查询 Option。
func withControllerIDsOption(ids interface{}) tgorm.Option {
	return tgorm.WithIn(db.ColumnControllerID, ids)
}

// withControllerIDOption 构造门禁控制器ID等值查询 Option。
func withControllerIDOption(id db.IDType) tgorm.Option {
	return tgorm.WithEqual(db.ColumnControllerID, id)
}

// withTimestampBetweenOption 构造时间戳范围查询 Option。
func withTimestampBetweenOption(begin, end int64) tgorm.Option {
	return tgorm.WithBetween(db.ColumnTimestamp, begin, end)
}

// withCreateTimeBetweenOption 构造创建时间范围查询 Option。
func withCreateTimeBetweenOption(begin, end int64) tgorm.Option {
	return tgorm.WithBetween(db.ColumnCreateTime, begin, end)
}

// withTimestampDescOption 构造时间戳降序排序 Option。
func withTimestampDescOption() tgorm.Option {
	return tgorm.WithDESC(db.ColumnTimestamp)
}

// withNotStaffType 构造排除指定人员类型的查询 Option。
func withNotStaffType(t db.StaffType) tgorm.Option {
	return tgorm.WithNotEqual(db.ColumnType, t)
}

// addMozuOptionIfNotEmpty 当模组ID非空时追加模组查询条件。
func addMozuOptionIfNotEmpty(opts []tgorm.Option, mozuID string) []tgorm.Option {
	if len(mozuID) > 0 {
		opts = append(opts, withMozuID(mozuID))
	}
	return opts
}

// addIDsOptionIfNotEmpty 当ID列表非空时追加ID查询条件。
func addIDsOptionIfNotEmpty(opts []tgorm.Option, requestIds []db.IDType) []tgorm.Option {
	if len(requestIds) > 0 {
		opts = append(opts, withIDsOption(requestIds))
	}
	return opts
}

// appendControllerIDsOptionIfNotEmpty 当控制器ID列表非空时追加查询条件。
func appendControllerIDsOptionIfNotEmpty(opts []tgorm.Option, controllerIDs []db.IDType) []tgorm.Option {
	if len(controllerIDs) > 0 {
		opts = append(opts, withControllerIDsOption(controllerIDs))
	}
	return opts
}

// appendCtrlDoorsOptIfNotEmpty 当控制器和门映射非空时追加查询条件。
// 特用于门禁控制器页面，门下刷卡记录按钮的查询。
func appendCtrlDoorsOptIfNotEmpty(opts []tgorm.Option, controllers map[int][]int) []tgorm.Option {
	// 特用于门禁控制器页面，门下刷卡记录按钮的查询，因此controller和door如果存在只会有一个。
	if len(controllers) > 0 {
		var controllerID, doorNum int
		for k, v := range controllers {
			controllerID = k
			doorNum = v[0]
			break
		}
		opts = append(opts, withControllerIDOption(controllerID))
		opts = append(opts, withDoorNumber(db.DoorNumberType(doorNum)))
	}
	return opts
}

// withCardNumbersOption 构造卡号列表查询 Option。
func withCardNumbersOption(cards []string) tgorm.Option {
	return tgorm.WithIn(db.ColumnCardNo, cards)
}

// withCardLike 构造卡号模糊查询 Option。
func withCardLike(card string) tgorm.Option {
	return tgorm.WithLike(db.ColumnCardNo, card)
}

// withCardNumberLike 构造卡号（card_number字段）模糊查询 Option。
func withCardNumberLike(card string) tgorm.Option {
	return tgorm.WithLike(db.ColumnCardNumber, card)
}

// withNameLike 构造名称模糊查询 Option。
func withNameLike(value string) tgorm.Option {
	return tgorm.WithLike(db.ColumnName, value)
}

// withUsernameLike 构造用户名模糊查询 Option。
func withUsernameLike(value string) tgorm.Option {
	return tgorm.WithLike(db.ColumnUsername, value)
}

// withJSONLike 构造 JSON 字段模糊查询 Option。
func withJSONLike(columnName string, fields []string, value string) tgorm.Option {
	return tgorm.WithJSONLike(columnName, fields, value)
}

// withPhoneLike 构造手机号模糊查询 Option。
func withPhoneLike(value string) tgorm.Option {
	return tgorm.WithLike(db.ColumnPhone, value)
}

// withEmailLike 构造邮箱模糊查询 Option。
func withEmailLike(value string) tgorm.Option {
	return tgorm.WithLike(db.ColumnEmail, value)
}

// withCommentLike 构造备注模糊查询 Option。
func withCommentLike(value string) tgorm.Option {
	return tgorm.WithLike(db.ColumnComment, value)
}

// withCompany 构造公司名称等值查询 Option。
func withCompany(value string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnCompany, value)
}

// addCompanyOptionIfNotEmpty 当公司名称非空时追加查询条件。
func addCompanyOptionIfNotEmpty(opts []tgorm.Option, value string) []tgorm.Option {
	if len(value) > 0 {
		opts = append(opts, withCompany(value))
	}
	return opts
}

// addDoorNameOptionIfNotEmpty 当门名称非空时追加查询条件。
func addDoorNameOptionIfNotEmpty(opts []tgorm.Option, doorName string) []tgorm.Option {
	if len(doorName) > 0 {
		opts = append(opts, withDoorName(doorName))
	}
	return opts
}

// withCardStaffIDs 构造人员ID列表查询 Option（用于卡片关联查询）。
func withCardStaffIDs(ids []db.IDType) tgorm.Option {
	return tgorm.WithIn(db.ColumnStaffID, ids)
}

// withCardFlag 构造卡片标志等值查询 Option。
func withCardFlag(f db.CardFlagType) tgorm.Option {
	return tgorm.WithEqual(db.ColumnCardFlag, f)
}

// withCardType 构造卡片类型等值查询 Option。
func withCardType(t db.CardType) tgorm.Option {
	return tgorm.WithEqual(db.ColumnCardType, t)
}

// withAccessGroupIDOption 构造权限组ID等值查询 Option。
func withAccessGroupIDOption(groupID db.IDType) tgorm.Option {
	return tgorm.WithEqual(db.ColumnAccessGroupID, groupID)
}

// withCardsMozuOption 构造卡号+模组ID组合查询条件列表。
func withCardsMozuOption(cards []string, mozuID string) []tgorm.Option {
	opts := make([]tgorm.Option, 0, 2)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = append(opts, withCardNumbersOption(cards))
	return opts
}

// withChannelID 构造通道ID等值查询 Option。
func withChannelID(channelID string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnChannelID, channelID)
}

// withChannelIDs 构造通道ID列表查询 Option。
func withChannelIDs(channelIDs []string) tgorm.Option {
	return tgorm.WithIn(db.ColumnChannelID, channelIDs)
}

// withTimestamp 构造时间戳等值查询 Option。
func withTimestamp(timestamp int64) tgorm.Option {
	return tgorm.WithEqual(db.ColumnTimestamp, timestamp)
}

// withState 构造告警状态等值查询 Option。
func withState(state db.AlarmStateType) tgorm.Option {
	return tgorm.WithEqual(db.ColumnState, state)
}

// withRequestState 构造请求状态等值查询 Option。
func withRequestState(state string) tgorm.Option { return tgorm.WithEqual(db.ColumnState, state) }

// withType 构造类型等值查询 Option。
func withType(alarmType int) tgorm.Option {
	return tgorm.WithEqual(db.ColumnType, alarmType)
}

// withDoorNumber 构造门编号等值查询 Option。
func withDoorNumber(doorNumber db.DoorNumberType) tgorm.Option {
	return tgorm.WithEqual(db.ColumnDoorNumber, doorNumber)
}

// withDoorName 构造门名称等值查询 Option。
func withDoorName(doorName string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnDoorName, doorName)
}

// withDirection 构造方向等值查询 Option。
func withDirection(direction int) tgorm.Option {
	return tgorm.WithEqual(db.ColumnDirection, direction)
}

// withStatuses 构造状态列表查询 Option。
func withStatuses(status []int) tgorm.Option {
	return tgorm.WithIn(db.ColumnStatus, status)
}

// withGroupNo 构造时间组编号等值查询 Option。
func withGroupNo(groupNo int) tgorm.Option {
	return tgorm.WithEqual(db.ColumnTimeGroupNo, groupNo)
}

// withoutGroupNo 构造排除指定时间组编号的查询 Option。
func withoutGroupNo(groupNo int) tgorm.Option {
	return tgorm.WithNotEqual(db.ColumnTimeGroupNo, groupNo)
}

// withCardNo 构造卡号等值查询 Option。
func withCardNo(cardNo string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnCardNo, cardNo)
}

// withCardNumber 构造卡号（card_number字段）等值查询 Option。
func withCardNumber(cardNumber string) tgorm.Option {
	return tgorm.WithEqual(db.ColumnCardNumber, cardNumber)
}

// queryAndCountDriverCards 分页查询驱动卡片记录并统计总数。
func queryAndCountDriverCards(tx *gorm.DB, offset int, limit int, values interface{}, count *int64) error {
	return tx.Offset(offset).Limit(limit).Find(values).Offset(-1).Count(count).Error
}

// queryAndCountDriverAlarms 分页查询驱动告警记录并统计总数。
func queryAndCountDriverAlarms(tx *gorm.DB, offset int, limit int, values interface{}, count *int64) error {
	return tx.Offset(offset).Limit(limit).Find(values).Offset(-1).Count(count).Error
}

// queryAndCountDriverEvents 分页查询驱动事件记录并统计总数。
func queryAndCountDriverEvents(tx *gorm.DB, offset int, limit int, values interface{}, count *int64) error {
	return tx.Offset(offset).Limit(limit).Find(values).Offset(-1).Count(count).Error
}

// withIndexDescOption 构造按索引降序排序 Option（index 为保留字，需反引号包裹）。
func withIndexDescOption() tgorm.Option {
	return tgorm.WithDESC("`" + db.ColumnIndex + "`")
}
