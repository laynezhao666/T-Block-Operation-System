// Package data 提供数据上报接口的抽象层。
package data

import (
	"context"

	"dac/entity/model/db"

	pb "dac/repo/pb/tcommon_point_data"
)

// writer 全局数据写入器实例
var (
	writer = newWriter()
)

// Writer 数据写入接口，定义点位数据上报方法。
type Writer interface {
	// SetTBOSPointsWithExtends 上报点位数据到TBOS平台
	SetTBOSPointsWithExtends(ctx context.Context,
		timestamp int64, kind pb.DataKind,
		extends map[string]string, code string,
		deviceID db.GIDType, points []*pb.Point) error
}

// GetWriter 获取全局数据写入器实例。
func GetWriter() Writer {
	return writer
}

// newWriter 创建数据写入器实例。
func newWriter() *writerImpl {
	return &writerImpl{}
}
