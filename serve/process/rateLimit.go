package process

import (
	"sync"
	"time"
)

const (
	rateWindow     = time.Minute //窗口长度
	rateMaxFails   = 5           //最大失败数
	rateCleanEvery = time.Minute //清理时间
)

type failRecord struct {
	windowStart time.Time //开始时间
	count       int       //失败次数
}

var (
	failMu  sync.Mutex
	failMap = make(map[string]*failRecord)
)

func allowed(ip string) bool {
	failMu.Lock()
	defer failMu.Unlock()
	rec, ok := failMap[ip]
	if !ok {
		//没有记录过
		return true
	} else {
		//在记录表中

		if time.Since(rec.windowStart) >= rateWindow {
			//超过一分钟了，过期重新开始计数
			delete(failMap, ip)
			return true
		} else {
			//一分钟内判断次数
			return rec.count < rateMaxFails
		}
	}
}

func recordFail(ip string) {
	failMu.Lock()
	defer failMu.Unlock()

	rec, ok := failMap[ip]
	if !ok || time.Since(rec.windowStart) >= rateWindow {
		//没记录或者记录已过期
		failMap[ip] = &failRecord{
			windowStart: time.Now(),
			count:       1,
		}
	} else {
		//有记录并且未过期
		rec.count++
	}

}

func clearFails(ip string) {
	failMu.Lock()
	defer failMu.Unlock()
	delete(failMap, ip)
}

func init() {
	//一个 init() 里跑每分钟清理的 ticker
	go func() {
		t := time.NewTicker(rateCleanEvery)
		defer t.Stop()

		for range t.C {
			failMu.Lock()
			for ip, rec := range failMap {
				if time.Since(rec.windowStart) >= rateWindow {
					delete(failMap, ip)
				}
			}
			failMu.Unlock()
		}
	}()
}
