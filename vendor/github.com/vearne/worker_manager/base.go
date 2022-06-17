package worker_manager

import (
	"sync/atomic"
)

const (
	True  = 1
	False = 0
)

type BoolFlag uint32

func NewBoolFlag(flag bool) *BoolFlag {
	var result BoolFlag
	if flag {
		result = BoolFlag(True)
	} else {
		result = BoolFlag(False)
	}
	return &result
}

func IsTrue(f *BoolFlag) bool {
	return atomic.LoadUint32((*uint32)(f)) == True
}

func SetFalse(f *BoolFlag) {
	atomic.CompareAndSwapUint32((*uint32)(f), True, False)
}

func SetTrue(f *BoolFlag) {
	atomic.CompareAndSwapUint32((*uint32)(f), False, True)
}
