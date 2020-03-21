package common

import "errors"

var (
	ERR_LOCK_BUSY   = errors.New("锁被占用!")
	ERR_NO_IP_FOUND = errors.New("机器没有物理网卡!")
)
