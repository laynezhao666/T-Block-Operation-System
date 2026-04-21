// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"dac/entity/consts"
	"dac/entity/model/db"
)

// HideIDCard 隐藏身份证号后四位，替换为掩码字符
func HideIDCard(idCard string) string {
	if len(idCard) < len(consts.Mask) {
		return idCard
	}
	return idCard[:len(idCard)-len(consts.Mask)] + consts.Mask
}

// ProcessPersonalInformation 处理敏感信息，隐藏指纹、密码和身份证号
func ProcessPersonalInformation(s *db.Staff) {
	if s == nil {
		return
	}

	s.Fingerprint = consts.HidedPasswordString
	// 隐藏密码
	s.Password = consts.HidedPasswordString
	// 隐藏身份证后四位
	s.Paper = HideIDCard(s.Paper)
}
