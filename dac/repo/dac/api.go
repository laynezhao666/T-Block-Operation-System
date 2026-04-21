package dac

import (
	"context"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"

	"gorm.io/gorm"
)

var (
	rw = newRW()
)

// RW 定义门禁数据访问层的读写接口，包含控制器、门、卡片、权限组、事件、告警等所有数据操作。
type RW interface {
	// Init 初始化数据库连接和表结构。
	Init() error

	// GetMozuWithSameBuildings 获取与指定模组同楼栋的所有模组。
	GetMozuWithSameBuildings(ctx context.Context, mozuID int) ([]db.Mozu, error)

	// AddGroup 添加门分组并关联门。
	AddGroup(ctx context.Context, group db.DoorGroup, doorIDs []db.IDType) error
	// UpdateGroup 更新门分组及其关联的门。
	UpdateGroup(ctx context.Context, group db.DoorGroup, doorIDs []db.IDType) error
	// DeleteGroup 删除门分组。
	DeleteGroup(ctx context.Context, id db.IDType) error
	// GetAllDoorGroups 获取指定模组下所有门分组。
	GetAllDoorGroups(ctx context.Context, mozuID string) ([]db.DoorGroup, error)
	// GetGroupDoors 获取指定分组下的所有门。
	GetGroupDoors(ctx context.Context, group db.IDType) ([]db.Door, error)

	// AddRequests 批量添加请求记录。
	AddRequests(ctx context.Context, reqs []db.Request) error
	// DeleteRequests 批量删除请求记录。
	DeleteRequests(ctx context.Context, ids []db.IDType) error
	// GetRequests 分页查询请求记录（支持多条件过滤）。
	GetRequests(ctx context.Context, mozuID string, offset int, limit int, query string,
		beginTime int64, endTime int64, queryCreateTime bool, state string, queryState bool,
		method string, queryMethod bool) (int64, []db.Request, error)
	// GetAllRequests 获取指定模组下所有请求记录。
	GetAllRequests(ctx context.Context, mozuID string) ([]db.Request, error)
	// CleanRequests 清理请求记录（通过回调函数决定保留哪些）。
	CleanRequests(ctx context.Context, afterGet func(*gorm.DB, []db.Request) ([]db.Request, error)) ([]db.Request, error)
	// GetRequestsByIds 根据ID列表获取请求记录。
	GetRequestsByIds(ctx context.Context, mozuID string, requestIds []db.IDType) ([]db.Request, error)
	// GetRequestsByControllers 获取指定控制器的所有请求记录。
	GetRequestsByControllers(ctx context.Context, controllerID db.IDType) ([]db.Request, error)
	// GetAllRequestWithControllerInfo 分页获取请求记录并附带控制器信息。
	GetAllRequestWithControllerInfo(ctx context.Context, mozuID string, offset int, limit int,
		query string, method string) (int64, []db.Request, map[db.IDType]ControllerInfo, error)
	// FetchRequests 获取指定数量的请求记录。
	FetchRequests(ctx context.Context, n int) ([]db.Request, error)
	// UpdateFailedRequests 更新请求失败状态及其消息。
	UpdateFailedRequests(ctx context.Context, failedIds []db.IDType, messages map[db.IDType]string) error
	// UpdateSuccessRequests 更新请求成功状态。
	UpdateSuccessRequests(ctx context.Context, successIds []db.IDType) error
	// UpdateRequestsInfo 更新请求的方法、载荷、创建时间和状态。
	UpdateRequestsInfo(ctx context.Context, ids []db.IDType, method, payload string, createTime int64, state string) error
	// BatchReExecuteRequestsInfo 批量重新执行请求（更新创建时间和状态）。
	BatchReExecuteRequestsInfo(ctx context.Context, ids []db.IDType, createTime int64, state string) error
	// OutdateRequests 标记请求为过时状态。
	OutdateRequests(ctx context.Context, ids []db.IDType) error
	// DeleteRequestsByTime 删除指定时间之前的请求记录。
	DeleteRequestsByTime(ctx context.Context, timeThreshold int64) (int64, error)
	// OutdatedRequestsByTime 获取指定时间之前的过时请求数量。
	OutdatedRequestsByTime(ctx context.Context, timeThreshold int64, currentState string) (int64, error)

	// GetAllDoorControllersAndDoors 获取指定模组下所有门禁控制器及其关联的门。
	GetAllDoorControllersAndDoors(ctx context.Context, mozuID string) ([]db.DoorController, map[db.IDType][]db.Door, error)
	// GetAllDoorControllers 获取指定模组下所有门禁控制器。
	GetAllDoorControllers(ctx context.Context, mozuID string) ([]db.DoorController, error)
	// UpdateDoorController 更新门禁控制器信息（支持前置和后置回调）。
	UpdateDoorController(ctx context.Context, id db.IDType, controller *db.DoorController,
		beforeUpdate func(tx *gorm.DB) error, afterUpdate func(tx *gorm.DB) error) error
	// AddDoorControllers 批量添加门禁控制器（支持前置和后置回调）。
	AddDoorControllers(ctx context.Context, controllers []rt.DoorController,
		beforeAdd func(*gorm.DB) error, afterAdd func(*gorm.DB) error) error
	// GetDoorControllerRecord 根据ID获取单个门禁控制器记录。
	GetDoorControllerRecord(ctx context.Context, id db.IDType) (db.DoorController, error)
	// GetDoorControllers 根据ID列表获取门禁控制器映射。
	GetDoorControllers(ctx context.Context, ids []db.IDType) (map[db.IDType]rt.DoorController, error)
	// GetControllerNames 根据ID列表获取控制器名称映射。
	GetControllerNames(ctx context.Context, ids []db.IDType) (map[db.IDType]string, error)
	// DeleteDoorControllers 批量删除门禁控制器。
	DeleteDoorControllers(ctx context.Context, ids []db.IDType) error

	// 更新门的名称和参数（支持前置和后置回调）。
	UpdateDoorsNameAndParams(ctx context.Context, ids []db.IDType,
		names map[db.IDType]string,
		params map[db.IDType]*db.DoorParameter,
		beforeUpdate func(*gorm.DB) error,
		afterUpdate func(*gorm.DB) error) error
	// UpdateDoor 更新单个门信息（支持前置和后置回调）。
	UpdateDoor(ctx context.Context, id db.IDType, d *db.Door,
		beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error
	// UpdateDoors 批量更新门信息（支持前置和后置回调）。
	UpdateDoors(ctx context.Context, doors map[db.IDType]*db.Door,
		beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error
	// UpdateDoorCode 更新门的采集编码（支持前置和后置回调）。
	UpdateDoorCode(ctx context.Context, id db.IDType, code string,
		beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error
	// SetDoors 设置控制器下的门列表（自动处理新增和删除）。
	SetDoors(ctx context.Context, controllerID db.IDType, doors []db.Door, afterSave func(*gorm.DB) error) error
	// GetDoors 根据ID列表获取门信息。
	GetDoors(ctx context.Context, ids []db.IDType) ([]db.Door, error)
	// GetDoor 根据ID获取单个门信息。
	GetDoor(ctx context.Context, id db.IDType) (db.Door, error)
	// GetAllDoors 获取所有门信息。
	GetAllDoors(ctx context.Context) ([]db.Door, error)
	// GetControllerDoors 获取指定控制器下的所有门（支持后置回调）。
	GetControllerDoors(ctx context.Context, controllerID db.IDType,
		afterGet func(*gorm.DB, []db.Door) error) ([]db.Door, error)

	// AddDefaultTimeGroups 添加默认时间组。
	AddDefaultTimeGroups(ctx context.Context) error
	// GetAllTimeGroups 获取所有时间组。
	GetAllTimeGroups(ctx context.Context) ([]db.TimeGroup, error)
	// GetTimeGroupsByNos 根据时间组编号列表获取时间组映射。
	GetTimeGroupsByNos(ctx context.Context, groupNos []int) (map[int]db.TimeGroup, error)
	// UpdateTimeGroup 更新时间组（支持前置和后置回调）。
	UpdateTimeGroup(ctx context.Context, timeGroup db.TimeGroup,
		beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error

	// GetAllStaffCompany 获取指定模组下所有人员的公司列表（去重排序）。
	GetAllStaffCompany(ctx context.Context, mozuID string) ([]string, error)
	// GetAllCardStaffMap 获取所有卡片与人员的映射关系（按模组分组）。
	GetAllCardStaffMap(ctx context.Context) (map[string][]db.Card, map[string]map[string]db.Staff, error)
	// GetStaffs 分页查询人员列表（支持模糊搜索和公司过滤）。
	GetStaffs(ctx context.Context, mozuID string, offset int, limit int,
		query string, company string) (int64, []db.Staff, error)
	// GetAllStaffs 获取指定模组下所有人员（排除已删除）。
	GetAllStaffs(ctx context.Context, mozuID string) ([]db.Staff, error)
	// GetStaffsByID 根据ID列表获取人员映射。
	GetStaffsByID(ctx context.Context, ids []db.IDType) (map[db.IDType]db.Staff, error)
	// AddStaff 添加单个人员记录，返回新建ID。
	AddStaff(ctx context.Context, staff db.Staff) (db.IDType, error)
	// AddStaffs 批量添加人员记录。
	AddStaffs(ctx context.Context, staffs []db.Staff) error
	// UpdateStaff 更新人员信息（支持前置和后置回调）。
	UpdateStaff(ctx context.Context, id db.IDType, staff db.Staff,
		beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error
	// DeleteStaff 软删除人员（标记为已删除，解绑关联卡片）。
	DeleteStaff(ctx context.Context, id db.IDType,
		beforeDelete func(*gorm.DB) error, afterDelete func(*gorm.DB) error) error

	// GetAllCards 获取所有卡片（支持后置回调过滤）。
	GetAllCards(ctx context.Context, afterGet func(*gorm.DB, []db.Card) ([]db.Card, error)) ([]db.Card, error)
	// GetAllCardsWithStaffAndAccessGroup 获取所有卡片及其关联的人员和权限组信息。
	GetAllCardsWithStaffAndAccessGroup(ctx context.Context,
		mozuID string,
	) ([]db.Card, map[db.IDType]db.Staff,
		map[string][]db.IDType, map[db.IDType]string, error)
	// GetCards 分页查询卡片列表（支持多条件过滤）。
	GetCards(ctx context.Context, mozuID string,
		cardNumbers []string, query string,
		cardType db.CardType, queryCardType bool,
		cardFlag db.CardFlagType, queryCardFlag bool,
		accessGroupID db.IDType, queryAccessGroup bool,
		offset int, limit int) (int64, []db.Card, error)
	// GetCardsByCardNos 根据卡号列表获取卡片信息。
	GetCardsByCardNos(ctx context.Context, cardNos []string, mozuID string) ([]db.Card, error)
	// GetCardsByStaffs 根据人员ID列表获取关联的卡片。
	GetCardsByStaffs(ctx context.Context, staffIDs []db.IDType) ([]db.Card, error)
	// AddCard 添加卡片记录（支持后置回调）。
	AddCard(ctx context.Context, card db.Card, afterAdd func(tx *gorm.DB) error) error
	// UpdateCardsFlag 批量更新卡片标志（支持后置回调）。
	UpdateCardsFlag(ctx context.Context, cards []string, mozuID string, flag db.CardFlagType,
		afterUpdate func(*gorm.DB) error) error
	// UpdateCardsType 批量更新卡片类型。
	UpdateCardsType(ctx context.Context, cards []string, mozuID string, cardType db.CardType) error
	// UpdateCardValidTime 批量更新卡片有效期（支持后置回调）。
	UpdateCardValidTime(ctx context.Context, cards []string, mozuID string, validTime int64,
		afterUpdate func(*gorm.DB) error) error
	// UpdateCardsStaff 批量更新卡片关联的人员（支持后置回调）。
	UpdateCardsStaff(ctx context.Context, cards []string, staffID db.IDType, mozuID string,
		afterUpdate func(tx *gorm.DB) error) error
	// DeleteCards 批量删除卡片及其权限组关联（支持前置回调）。
	DeleteCards(ctx context.Context, cards []string, mozuID string, beforeDelete func(*gorm.DB) error) error
	// UnbindCards 批量解绑卡片与人员的关联。
	UnbindCards(ctx context.Context, cards []string, mozuID string) error

	// GetCardAccessGroupRelationByCards 获取卡片与权限组的关联关系。
	GetCardAccessGroupRelationByCards(ctx context.Context, cards []string, mozuID string) ([]db.CardAccessRelation, error)
	// UpdateCardAccessGroupRelation 更新卡片与权限组的关联关系（支持前置和后置回调）。
	UpdateCardAccessGroupRelation(ctx context.Context, cards []string, groups []db.IDType, mozuID string,
		deleteIfEmpty bool, beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error

	// GetAllCardAccessGroups 获取指定模组下所有卡片权限组。
	GetAllCardAccessGroups(ctx context.Context, mozuID string) ([]db.AccessGroup, error)
	// GetAccessGroupsBaseInfo 根据ID列表获取权限组基本信息。
	GetAccessGroupsBaseInfo(ctx context.Context, ids []db.IDType) (map[db.IDType]db.AccessGroupBaseInfo, error)
	// GetAccessGroups 分页获取权限组列表。
	GetAccessGroups(ctx context.Context, mozuID string, offset int, limit int) (int64, []db.AccessGroup, error)
	// GetAllAccessGroups 获取指定模组下所有权限组。
	GetAllAccessGroups(ctx context.Context, mozuID string) ([]db.AccessGroup, error)
	// GetAccessGroupsByID 根据ID列表获取权限组映射。
	GetAccessGroupsByID(ctx context.Context, ids []db.IDType) (map[db.IDType]db.AccessGroup, error)
	// GetAccessGroupDoors 获取权限组关联的门列表。
	GetAccessGroupDoors(ctx context.Context, ids []db.IDType) (map[db.IDType][]db.Door, error)
	// AddAccessGroup 添加权限组及其关联的门和卡片（支持前置和后置回调）。
	AddAccessGroup(ctx context.Context, accessGroupWrapper db.AccessGroupInfoWrapper, mozuID string,
		beforeAdd func(*gorm.DB) error, afterAdd func(*gorm.DB) error) (db.IDType, error)
	// UpdateAccessGroup 更新权限组及其关联关系（支持前置和后置回调）。
	UpdateAccessGroup(ctx context.Context, id db.IDType, accessGroupWrapper db.AccessGroupInfoWrapper,
		mozuID string, beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error
	// DeleteAccessGroup 删除权限组及其关联关系（支持前置和后置回调）。
	DeleteAccessGroup(ctx context.Context, id db.IDType,
		beforeDelete func(*gorm.DB) error, afterDelete func(*gorm.DB) error) error
	// GetCardAndStaffsByAccessGroupIDs 根据权限组ID获取关联的卡片和人员信息。
	GetCardAndStaffsByAccessGroupIDs(ctx context.Context,
		ids []db.IDType, mozuID string,
	) (map[db.IDType][]db.CardAndStaffBase, error)

	// GetEventsByDoors 获取指定门的刷卡记录，controllerDoors 中 key 为门禁控制器 id，value 为门编号列表
	GetEventsByDoors(ctx context.Context, controllerDoors map[db.IDType][]int,
		offset, limit int, beginTime, endTime int64,
		afterGet func(*gorm.DB, []db.Event) error) (int64, []db.Event, error)
	GetEventsNumber(ctx context.Context, mozuID string, doorName string, controllerDoors map[int][]int,
		beginTime, endTime int64, query string, afterGet func(*gorm.DB, int64) error) (int64, error)
	GetEvents(ctx context.Context, mozuID string, controllerIDs []db.IDType, query string, doorName string,
		offset, limit int, beginTime, endTime int64, afterGet func(*gorm.DB, []db.Event) error) (int64, []db.Event, error)
	AddEvents(ctx context.Context, controllerID db.IDType, events []db.Event,
		checkExist bool, beginTime int64, endTime int64,
		fillEvent func(e *db.Event), afterAdd func(*gorm.DB) error) error

	// GetOrCreateEventIndex 获取或创建事件索引记录。
	GetOrCreateEventIndex(ctx context.Context, controllerID db.IDType) (db.EventIndexRecord, error)
	// GetOrCreateEventTimestampIndex 获取或创建事件时间戳索引记录。
	GetOrCreateEventTimestampIndex(ctx context.Context,
		controllerID db.IDType, mozuID string,
	) (db.EventTimestampIndexRecord, error)
	// UpdateEventIndex 更新事件索引。
	UpdateEventIndex(ctx context.Context, controllerID db.IDType, index, last int) error

	// GetAlarmsNumber 获取告警总数（支持后置回调）。
	GetAlarmsNumber(ctx context.Context, mozuID string,
		controllerIDs []db.IDType, beginTime, endTime int64,
		afterGet func(*gorm.DB, int64) error) (int64, error)
	GetAlarms(ctx context.Context, mozuID string, controllerIDs []db.IDType,
		offset, limit int, beginTime, endTime int64, afterGet func(*gorm.DB, []db.Alarm) error) (int64, []db.Alarm, error)
	AddAlarms(ctx context.Context, controllerID db.IDType, alarms []db.Alarm,
		checkExist bool, beginTime int64, endTime int64,
		fillAlarm func(*db.Alarm), afterAdd func(*gorm.DB) error) error

	// GetOrCreateAlarmIndex 获取或创建告警索引记录。
	GetOrCreateAlarmIndex(ctx context.Context, controllerID db.IDType) (db.AlarmIndexRecord, error)
	// GetOrCreateAlarmTimestampIndex 获取或创建告警时间戳索引记录。
	GetOrCreateAlarmTimestampIndex(ctx context.Context,
		controllerID db.IDType, mozuID string,
	) (db.AlarmTimestampIndexRecord, error)
	// UpdateAlarmIndex 更新告警索引。
	UpdateAlarmIndex(ctx context.Context, controllerID db.IDType, index, last int) error

	// UpdateControllerAndDoorGIDsByCode 根据采集编码批量更新控制器和门的GID。
	UpdateControllerAndDoorGIDsByCode(ctx context.Context, codeGIDs rt.CodeGIDMapType) error

	// GetDriverAllCards 获取指定控制器和通道下所有驱动卡片。
	GetDriverAllCards(ctx context.Context, controllerID db.IDType, channelID string, status []int) ([]db.DriverCard, error)
	// AddDriverCard 添加驱动卡片（自动判断新建或更新已有记录）。
	AddDriverCard(ctx context.Context, controllerID db.IDType, channelID string, card driver.Card,
		addCardByNewCardFunc func(tx *gorm.DB, card driver.Card) error,
		addCardByUpdateOldCardFunc func(tx *gorm.DB, card driver.Card, oldCard db.DriverCard) error) error
	// GetDriverCard 获取指定控制器、通道和卡号的驱动卡片。
	GetDriverCard(ctx context.Context, controllerID db.IDType, channelID string, cardNo string) (db.DriverCard, error)
	// LogicDeleteDriverCard 逻辑删除驱动卡片（标记状态为已删除）。
	LogicDeleteDriverCard(ctx context.Context, controllerID db.IDType, channelID string, cardNo string,
		f func(driverCard db.DriverCard) error) error
	// GetDriverCards 分页获取驱动卡片列表。
	GetDriverCards(ctx context.Context, controllerID db.IDType,
		channelID string, offset int, limit int,
		status []int) (int64, []db.DriverCard, error)
	// GetLastDriverCard 获取指定控制器和通道下最后一张驱动卡片。
	GetLastDriverCard(ctx context.Context, controllerID db.IDType, channelID string) (db.DriverCard, error)

	// AddDriverDoorParameter 添加单个驱动门参数。
	AddDriverDoorParameter(ctx context.Context, controllerID db.IDType, channelID string,
		driverDoorParameter db.DriverDoorParameter) error
	// AddDriverDoorParameters 批量添加驱动门参数。
	AddDriverDoorParameters(ctx context.Context, controllerID db.IDType, channelID string,
		driverDoorParameters []db.DriverDoorParameter) error
	// GetDriverDoorParameters 获取指定控制器和通道下所有驱动门参数。
	GetDriverDoorParameters(ctx context.Context,
		controllerID db.IDType, channelID string,
	) ([]db.DriverDoorParameter, error)

	// AddDriverTimeGroup 添加驱动时间组（冲突时更新）。
	AddDriverTimeGroup(ctx context.Context, driverTimeGroup db.DriverTimeGroup) error
	// GetDriverTimeGroup 获取指定控制器、通道和编号的驱动时间组。
	GetDriverTimeGroup(ctx context.Context,
		controllerID db.IDType, channelID string,
		groupNo int) (db.DriverTimeGroup, error)
	// ClearDriverTimeGroup 清除指定时间组编号的驱动时间组。
	ClearDriverTimeGroup(ctx context.Context,
		controllerID db.IDType, channelID string,
		timeGroupNo int,
		f func(dbTimeGroups []db.DriverTimeGroup) error) error

	// SetDriverEvents 批量写入驱动事件（自动递增索引）。
	SetDriverEvents(ctx context.Context, controllerID db.IDType, items []db.DriverEvent) error
	// GetDriverEvents 分页获取驱动事件列表。
	GetDriverEvents(ctx context.Context,
		controllerID db.IDType, channelID string,
		offset int, limit int) (int64, []db.DriverEvent, error)
	// GetDriverEvent 根据条件获取单条驱动事件。
	GetDriverEvent(ctx context.Context, controllerID db.IDType, driverEvent db.DriverEvent) (db.DriverEvent, error)
	// SetDriverAlarms 批量写入驱动告警（自动递增索引）。
	SetDriverAlarms(ctx context.Context, controllerID db.IDType, items []db.DriverAlarm) error
	// GetDriverAlarms 分页获取驱动告警列表。
	GetDriverAlarms(ctx context.Context,
		controllerID db.IDType, channelID string,
		offset int, limit int) (int64, []db.DriverAlarm, error)
	// GetDriverAlarm 根据条件获取单条驱动告警。
	GetDriverAlarm(ctx context.Context, controllerID db.IDType, driverAlarm db.DriverAlarm) (db.DriverAlarm, error)

	// Transaction 在事务中执行操作。
	Transaction(ctx context.Context, f func(tx *gorm.DB) error) error
}

// GetRW 获取全局数据访问层实例。
func GetRW() RW {
	return rw
}

// IsInitialized 检查数据库是否已初始化
func IsInitialized() bool {
	if impl, ok := rw.(*impl); ok {
		return impl.db != nil
	}
	return false
}

// newRW 创建数据访问层实例。
func newRW() RW {
	return &impl{}
}
