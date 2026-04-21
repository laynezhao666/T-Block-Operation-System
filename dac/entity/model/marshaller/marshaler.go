// Package marshaller 提供门禁协议的TCP数据序列化和反序列化接口。
package marshaller

// TcpMarshal TCP协议数据的序列化/反序列化接口。
// Marshal 将命令和数据序列化为字节流。
// Unmarshal 将字节流反序列化为对应命令的数据结构。
type TcpMarshal interface {
	Marshal(cmd uint32, data interface{}) ([]byte, error)
	Unmarshal(cmd uint32, data []byte) (interface{}, error)
}
