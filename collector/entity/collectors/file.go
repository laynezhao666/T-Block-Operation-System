package collectors

import (
	"path/filepath"
)

const (
	sep                       string = string(filepath.Separator)
	CollectorConfigDir        string = "." + sep + "conf" + sep + "collect" + sep
	StdDevicesConfigDir       string = CollectorConfigDir + "std_devices" + sep
	CollectDevicesConfigDir   string = CollectorConfigDir + "collect_devices" + sep
	StdPointsConfigDir        string = CollectorConfigDir + "std_points" + sep
	ConfigModifyTimeDir       string = CollectorConfigDir + "modify_time" + sep
	CollectTemplatesConfigDir string = CollectorConfigDir + "collect_templates" + sep
	JsonFileSuffix            string = ".json"
)
