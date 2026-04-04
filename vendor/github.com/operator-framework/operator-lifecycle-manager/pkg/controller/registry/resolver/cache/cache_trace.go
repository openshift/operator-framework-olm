package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func cacheTraceBegin(log logrus.StdLogger, phase string, kvs ...string) time.Time {
	log.Printf("RESOLUTION_TRACE phase=%s event=BEGIN %s", phase, strings.Join(kvs, " "))
	return time.Now()
}

func cacheTraceDone(log logrus.StdLogger, phase string, start time.Time, kvs ...string) {
	ms := time.Since(start).Milliseconds()
	log.Printf("RESOLUTION_TRACE phase=%s event=DONE duration_ms=%d %s", phase, ms, strings.Join(kvs, " "))
}

func cacheKV(key string, value interface{}) string {
	return fmt.Sprintf("%s=%v", key, value)
}
