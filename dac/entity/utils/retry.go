// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"time"
)

// Retry 带指数退避的重试机制。
// retryNumber: 最大重试次数；maxWaitTime: 最大等待时间。
// f: 待重试的函数；successFun: 成功回调；
// retryFun: 每次重试回调；failedFun: 最终失败回调。
func Retry(retryNumber int, maxWaitTime time.Duration,
	f func() error, successFun func(),
	retryFun func(error), failedFun func(error),
) {
	var err error
	waitTime := time.Second
	for i := 0; i < retryNumber; i++ {
		if err = f(); err == nil {
			if successFun != nil {
				successFun()
			}
			return
		}

		if retryFun != nil {
			retryFun(err)
		}

		// 指数退避，每次等待时间翻倍
		waitTime <<= 1
		if waitTime > maxWaitTime {
			waitTime = maxWaitTime
		}
	}

	if failedFun != nil {
		failedFun(err)
	}
}
