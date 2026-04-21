// Package db 定义门禁系统的数据库模型和表结构。
package db

// Request 门禁控制器请求数据库模型，用于异步任务调度
type Request struct {
	ID           IDType `json:"id" gorm:"primaryKey;autoIncrement"`               // 主键
	ControllerID IDType `json:"controller_id" gorm:"column:controller_id;index"`  // 控制器ID
	Method       string `json:"method"`                                           // 请求方法
	Payload      []byte `json:"payload"`                                          // 请求载荷
	Message      string `json:"message" gorm:"column:message"`                    // 响应消息
	CreateTime   int64  `json:"create_time" gorm:"column:create_time;index"`      // 创建时间
	AccessTime   int64  `json:"access_time" gorm:"column:access_time;index"`      // 访问时间
	MozuID       string `json:"mozu_id" gorm:"column:mozu_id;index"`              // 模组ID
	State        string `json:"state" gorm:"index;column:state;type:varchar(10)"` // 请求状态
}

// TableName 返回请求表名
func (Request) TableName() string {
	return "t_dac_request"
}
