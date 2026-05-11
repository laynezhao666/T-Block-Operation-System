package cm

import (
	"agent/entity/config"
	"agent/repo/cm"
	"errors"
)

// Init 初始化
func Init() error {
	source := config.GetRB().Project.Source
	switch source {
	case cm.LocalFileConfigModName:
		return Worker().Init(source, nil)
	case cm.TaskServerModName:
		return Worker().Init(source, nil)
	case cm.TLinkModName:
		return Worker().Init(source, nil)
	}

	return errors.New(" cm init err")

}
