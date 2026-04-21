// Package rt 定义门禁系统的实时数据模型。
package rt

import (
	"dac/entity/model/db"
)

// DoorWithCodeItem 门信息导入导出项，包含门名称、编号、控制器IP和编码
type DoorWithCodeItem struct {
	Name         string `xlsx:"0"` // 门名称
	Number       int    `xlsx:"1"` // 门编号
	ControllerIP string `xlsx:"3"` // 控制器IP地址
	Code         string `xlsx:"4"` // IDCDB编码
	Extend       string `xlsx:"5"` // 扩展属性
}

// CodeGIDMapType 编码到GID的映射类型
type CodeGIDMapType map[string]db.GIDType
