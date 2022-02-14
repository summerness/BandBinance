package util

import (
	"errors"
	"fmt"
	"net"
	"time"
)

func Retry(tryTimes int, sleep float64, callback func() error) error {
	var res = errors.New("重试未知错误")

	for i := 1; i <= tryTimes; i++ {
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
		if i == tryTimes {
			return errors.New(fmt.Sprintf("重试超过次数: %d", tryTimes))
		}
		time.Sleep(time.Duration(sleep))
	}

	return res
}
