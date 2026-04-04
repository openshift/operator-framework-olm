package resolver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func NewResolutionID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// WithLog returns a shallow copy of the Resolver with the given logger.
// Use this to propagate per-resolution fields (e.g. resolution_id) into
// inner resolution layers without mutating the shared Resolver.
func (r *Resolver) WithLog(log logrus.FieldLogger) *Resolver {
	cp := *r
	cp.log = log
	return &cp
}

func traceBegin(log logrus.FieldLogger, phase string, kvs ...string) time.Time {
	if log != nil {
		log.Infof("RESOLUTION_TRACE phase=%s event=BEGIN %s", phase, strings.Join(kvs, " "))
	}
	return time.Now()
}

func traceDone(log logrus.FieldLogger, phase string, start time.Time, kvs ...string) {
	if log != nil {
		ms := time.Since(start).Milliseconds()
		log.Infof("RESOLUTION_TRACE phase=%s event=DONE duration_ms=%d %s", phase, ms, strings.Join(kvs, " "))
	}
}

// stdTraceBegin is like traceBegin but for logrus.StdLogger (Printf-only).
func stdTraceBegin(log logrus.StdLogger, phase string, kvs ...string) time.Time {
	if log != nil {
		log.Printf("RESOLUTION_TRACE phase=%s event=BEGIN %s", phase, strings.Join(kvs, " "))
	}
	return time.Now()
}

// stdTraceDone is like traceDone but for logrus.StdLogger (Printf-only).
func stdTraceDone(log logrus.StdLogger, phase string, start time.Time, kvs ...string) {
	if log != nil {
		ms := time.Since(start).Milliseconds()
		log.Printf("RESOLUTION_TRACE phase=%s event=DONE duration_ms=%d %s", phase, ms, strings.Join(kvs, " "))
	}
}

func kv(key string, value interface{}) string {
	return fmt.Sprintf("%s=%v", key, value)
}
