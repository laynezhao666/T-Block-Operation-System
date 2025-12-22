package cm

import (
	"errors"
	"agent/entity/config"
	"agent/repo/cm"
)
// Init 初始化
func Init() error {
	source := config.GetRB().Project.Source
	switch source {
	case cm.LocalFileConfigModName:
		return Worker().Init(source)
	case cm.TaskServerModName:
		return Worker().Init(source)
	case cm.TLinkModName:
		return Worker().Init(source)
	}

	return errors.New(" cm init err")

}
