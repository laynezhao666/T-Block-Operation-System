package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var (
	workerID string
)

func init() {
	// 初始化随机种子
	rand.Seed(time.Now().Unix())

	workerID = GetNewUUID()
}

// WorkerID 获取workerID
func WorkerID() string {
	return workerID
}

// GetNewUUID 获取新的UUID
func GetNewUUID() string {
	u, e := uuid.NewUUID()
	if e != nil {
		return fmt.Sprintf("%v", rand.Uint64())
	}
	return u.String()
}
