package simpleRetry

import (
	log "github.com/sirupsen/logrus"
	"strconv"
)

type retryFunc func(retryTimes int) (retry bool)
type simpleRetryFunc func() (err error)

var emit = func(tagKV map[string]string) {}

func doSync(name string, times int, run retryFunc) (limited bool) {
	if times <= 0 {
		return
	}

	var metricsFlagNeedRetry, metricsFlagFinalPass, metricsFlagLimited bool

	counter := getMetrics(name)
	for i := 0; i < times; i++ {
		if i > 0 && !counter.shouldRetry() {
			limited = true
			metricsFlagLimited = true
			break
		}

		if !run(i) {
			metricsFlagFinalPass = true
			counter.success.Increment(1)
			break
		}
		metricsFlagNeedRetry = true
		counter.fail.Increment(1)
	}

	// record by metrics
	emit(map[string]string{
		"name":             name,
		"need_retry":       strconv.FormatBool(metricsFlagNeedRetry),
		"final_pass":       strconv.FormatBool(metricsFlagFinalPass),
		"should_not_retry": strconv.FormatBool(metricsFlagLimited),
	})
	return
}
func SimpleDoSync(name string, times int, run simpleRetryFunc) (limited bool) {
	if times <= 0 {
		log.WithFields(log.Fields{"name": name}).Error("[retry.SimpleDoSync] CANNOT accept input times <= 0")
		return
	}

	limited = doSync(name, times, func(tryingTimes int) (retry bool) {
		err := run()
		if err != nil {
			log.WithFields(log.Fields{"name": name, "tryTimes": tryingTimes, "error": err}).Warn("[retry.SimpleDoSync] need retry")
			return true
		}
		return false
	})
	if limited {
		log.WithFields(log.Fields{"name": name}).Warn("[retry.SimpleDoSync] got limited")
	}

	return
}
