package simpleRetry

import (
	"sync"
	"time"
)

const (
	retryThreshold = 30
)

type metrics struct {
	success *Number
	fail    *Number
	tagKV   map[string]string
}

func (m *metrics) shouldRetry() bool {
	now := time.Now()
	failedTimes := m.fail.Sum(now)
	if failedTimes > retryThreshold && failedTimes > 0.1*m.success.Sum(now) {
		return false
	}
	return true
}

var retryMetrics = make(map[string]*metrics)
var retryMetricsLock sync.RWMutex

func getMetrics(name string) *metrics {
	retryMetricsLock.RLock()
	if m, ok := retryMetrics[name]; ok {
		retryMetricsLock.RUnlock()
		return m
	}
	retryMetricsLock.RUnlock()

	retryMetricsLock.Lock()
	defer retryMetricsLock.Unlock()

	var ok bool
	var m *metrics

	if m, ok = retryMetrics[name]; !ok {
		m = &metrics{
			success: NewNumber(),
			fail:    NewNumber(),
			tagKV:   map[string]string{"name": name},
		}
		retryMetrics[name] = m
	}

	return m
}
