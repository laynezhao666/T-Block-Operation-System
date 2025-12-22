package distributor

import "time"

const (
	DefaultHeartbeatInterval time.Duration = 10 * time.Second
)
// ClientConfig 客户端配置
type ClientConfig struct {
	Name              string `json:"name"`
	Target            string `json:"target"`
	Type              string `json:"type"`
	ClientId          string `json:"client_id"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
}
