package ratelimiter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Limiter Leaky bucket alogorithm based limiter
type Limiter struct {
	limit int
}

var limiters = make(map[string]*Limiter, 1)

// ParseLimitString parse limit string into maxLimit, fillLimit & fillRate
func ParseLimitString(limit string) (maxLimit, fillLimit int, fillRate time.Duration, err error) {
	tokens := strings.Split(limit, "-")
	if len(tokens) != 3 {
		err = fmt.Errorf("[%s] is not a valid limit", limit)
		return
	}
	maxLimit, err1 := strconv.Atoi(tokens[0])
	if err1 != nil {
		err = fmt.Errorf("[%v] not a valid max limit", tokens[0])
	}
	fillLimit, err1 = strconv.Atoi(tokens[1])
	if err1 != nil {
		err = fmt.Errorf("[%v] not a valid fill limit", tokens[0])
	}
	rate, err1 := strconv.Atoi(tokens[2])
	if err1 != nil {
		err = fmt.Errorf("[%v] not a valid fill rate", tokens[2])
	}
	fillRate = time.Duration(rate) * time.Millisecond

	return
}

// New new
func New(limit string) *Limiter {
	//parse limit string
	maxlimit, fillLimit, fillRate, _ := ParseLimitString(limit)

	lbucket := &Limiter{
		limit: maxlimit,
	}
	// start fill timer
	fillTicker := time.NewTicker(fillRate)
	go func() {
		for range fillTicker.C {
			newLimit := lbucket.limit + fillLimit
			if newLimit > maxlimit {
				newLimit = maxlimit
			}
			lbucket.limit = newLimit
		}
	}()

	return lbucket
}

// Consume Consume
func (lb *Limiter) Consume(duration int) (int, error) {
	if duration > lb.limit {
		err := fmt.Errorf("available limit[%v] not sufficient to consume[%v]", lb.limit, duration)
		lb.limit = 0
		return 0, err
	}

	// consume
	lb.limit = lb.limit - duration
	return lb.limit, nil
}

// AvailableLimit AvailableLimit
func (lb *Limiter) AvailableLimit() int {
	return lb.limit
}
