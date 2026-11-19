package utils

import (
	"fmt"
	"sync/atomic"
)

var msgSeq atomic.Uint64

func CreatMesId(userId int) string {
	return fmt.Sprintf("%d-%d", userId, msgSeq.Add(1))
}
