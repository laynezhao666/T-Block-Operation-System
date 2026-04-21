package rt

// StaffImportItem Excel导入人员信息的结构体
type StaffImportItem struct {
	Name      string `xlsx:"0"` // 姓名
	Password  string `xlsx:"1"` // 密码
	Sex       string `xlsx:"2"` // 性别
	Phone     string `xlsx:"3"` // 电话
	Email     string `xlsx:"4"` // 邮箱
	Company   string `xlsx:"5"` // 人员组
	PaperType string `xlsx:"6"` // 证件类型
	Paper     string `xlsx:"7"` // 证件号码
	Comment   string `xlsx:"8"` // 备注
}
