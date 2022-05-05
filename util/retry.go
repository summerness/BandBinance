package util

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// 用于失败时重新尝试运行 指定重试次数和休眠时间
func Retry(tryNumber int, sleep float64, callback func() error) error {
	var res = errors.New("重试未知错误")

	for i := 1; i <= tryNumber; i++ {
		err := callback()
		if err == nil {
			return nil
		}
		// 只有网络异常才重试
		_, ok := err.(net.Error)
		if !ok {
			res = err
			break
		}
		if i == tryNumber {
			return errors.New(fmt.Sprintf("重试超过次数: %d", tryNumber))
		}
		time.Sleep(time.Duration(sleep))
	}

	return res
}
