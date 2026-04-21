// Package db 定义门禁系统的数据库模型和表结构。
package db

// 数据库表列名常量，用于构建查询条件
const (
	ColumnID                     = "id"                       // 主键ID
	ColumnCode                   = "code"                     // 编码
	ColumnGID                    = "gid"                      // 全局ID
	ColumnName                   = "name"                     // 名称
	ColumnUsername               = "username"                 // 用户名
	ColumnMozuID                 = "mozu_id"                  // 模组ID
	ColumnNumber                 = "number"                   // 编号
	ColumnParameters             = "parameters"               // 参数
	ColumnControllerID           = "controller_id"            // 控制器ID
	ColumnTimestamp              = "timestamp"                // 时间戳
	ColumnHistorySyncedTimestamp = "history_synced_timestamp" // 历史同步时间戳
	ColumnCurrentSyncedTimestamp = "current_synced_timestamp" // 当前同步时间戳
	ColumnCreateTime             = "create_time"              // 创建时间
	ColumnIndex                  = "index"                    // 索引
	ColumnLast                   = "last"                     // 最后位置
	ColumnUpdateTime             = "update_time"              // 更新时间
	ColumnType                   = "type"                     // 类型
	ColumnCardNo                 = "card_no"                  // 卡号
	ColumnStaffID                = "staff_id"                 // 人员ID
	ColumnCardFlag               = "card_flag"                // 卡状态标志
	ColumnCardType               = "card_type"                // 卡类型
	ColumnCardValidTime          = "valid_time"               // 卡有效期
	ColumnAccessGroupID          = "access_group_id"          // 权限组ID
	ColumnPhone                  = "phone"                    // 电话
	ColumnEmail                  = "email"                    // 邮箱
	ColumnCompany                = "company"                  // 公司
	ColumnComment                = "comment"                  // 备注
	ColumnBuildingMID            = "building_mid"             // 楼栋MID
	ColumnChannel                = "channel"                  // 通道
	ColumnMethod                 = "method"                   // 方法
	ColumnChannelID              = "chid"                     // 通道ID
	ColumnTimeGroupNo            = "group_no"                 // 时间组编号
	ColumnStatus                 = "status"                   // 状态
	ColumnState                  = "state"                    // 门状态
	ColumnDoorNumber             = "door_number"              // 门编号
	ColumnDoorName               = "door_name"                // 门名称
	ColumnDirection              = "direction"                // 方向
	ColumnDoor                   = "door"                     // 门
	ColumnUserName               = "user_name"                // 用户名
	ColumnPassword               = "password"                 // 密码
	ColumnCardIndex              = "card_index"               // 卡索引
	ColumnCardNumber             = "card_number"              // 卡编号
)
