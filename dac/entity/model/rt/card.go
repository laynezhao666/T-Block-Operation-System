package rt

// CardItem Excel导入人员信息的结构体
type CardItem struct {
	Number    string `xlsx:"0"` // 卡号
	Type      string `xlsx:"1"` // 卡类型
	Flag      string `xlsx:"2"` // 卡状态
	ValidTime string `xlsx:"3"` // 有效期
	StaffName string `xlsx:"4"` // 人员名称
	StaffID   string `xlsx:"5"` // 人员编号
	Access    string `xlsx:"6"` // 权限组
	MozuID    string `xlsx:"7"` // 模组ID
}
