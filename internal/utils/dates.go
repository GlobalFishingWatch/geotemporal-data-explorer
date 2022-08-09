package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
)

func ParseDateRange(dateRange string) (time.Time, time.Time, error) {
	parts := strings.Split(dateRange, ",")
	if !strings.Contains(parts[0], "T") {
		parts[0] = fmt.Sprintf("%sT00:00:00.000Z", parts[0])
	}
	startRange, err := time.Parse("2006-01-02T15:04:05.999Z", parts[0])
	if err != nil {
		log.Errorf("From date not valid (%s)\n", parts[0])
		return time.Now(), time.Now(), types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "date-range",
			Detail: fmt.Sprintf("Date (%s) not valid", parts[0]),
		}})
	}
	if !strings.Contains(parts[1], "T") {
		parts[1] = fmt.Sprintf("%sT00:00:00.000Z", parts[1])
	}
	endRange, err := time.Parse("2006-01-02T15:04:05.999Z", parts[1])
	if err != nil {
		log.Errorf("To date not valid (%s)\n", parts[1])
		return time.Now(), time.Now(), types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "date-range",
			Detail: fmt.Sprintf("Date (%s) not valid", parts[1]),
		}})
	}
	return startRange, endRange, nil
}

func MonthDiff(a, b time.Time) int {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year := int(y2 - y1)
	month := int(M2 - M1)
	if d2 <= d1 {
		month = month - 1
	}

	return year*12 + month + 1
}

func DaysDiff(a, b time.Time) int {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	return int((b.UnixMilli() - a.UnixMilli()) / (24 * 60 * 60 * 1000))
}
