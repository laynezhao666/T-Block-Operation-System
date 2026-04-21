// Package rt 定义门禁系统运行时数据模型。
package rt

// PingArgs 控制器连通性测试参数
type PingArgs struct {
	Host            string `json:"host" binding:"required"`    // 控制器IP地址
	ProtocolName    string `json:"protocol_name"`              // 协议名称
	ProtocolVersion string `json:"protocol_version"`           // 协议版本
	Timeout         string `json:"timeout" binding:"required"` // 超时时间
	Account         string `json:"account"`                    // 账号
	Password        string `json:"password"`                   // 密码
}
