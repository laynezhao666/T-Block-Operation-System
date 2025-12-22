package utils

var (
	hostName = "unknown hostname"
)

func init() {
	//h, err := os.Hostname()
	//if err == nil {
	//	hostName = h
	//}
}

// GetHostName 获取主机名
func GetHostName() string {
	return hostName
}
