package consts

const (
	ConfigFile = "conf/agent.json"
	TRPCFile   = "./trpc_go.yaml"

	ModuleGroupRegexPrefix = "regex:"
	ModuleGroupRegex       = ModuleGroupRegexPrefix + "-(\\d+)-"

	ProjectPath          = "project"
	DeployPath           = "deploy"
	RelativeDeviceFile   = "devices.json"
	RelativeOptionFile   = "conf/option.json"
	RelativeTemplateFile = "templates.json"
	RelativeTemplateDir  = "templates"
	RelativeStdFile      = "std.json"
	StdTag               = "std"
	DeviceTag            = "devices"
	SuffixJSON           = ".json"
	StdDeviceFile        = "std_device.json"
	StdDeviceTag         = "std_device"
	EmptyDevicesXlsx     = "template_for_collect_devices.xlsx"
	EmptyTemplatesXlsx   = "template_for_collect_templates.xlsx"
	RelativeExcelPath    = "excel"
)
