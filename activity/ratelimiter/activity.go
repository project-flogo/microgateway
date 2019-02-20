package ratelimiter

import (
	"context"
	"sync"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
)

const (
	// MemorySize is the size of the circular buffer holding the request times
	MemorySize = 256
)

var (
	activityMetadata = activity.ToMetadata(&Settings{}, &Input{}, &Output{})
)

func init() {
	activity.Register(&Activity{}, New)
}

// Context is a token context
type Context struct {
	sync.Mutex
	index, prev, size int
	memory            [MemorySize]int64
}

// Activity is a rate limiter service
// Limit can be specified in the format "<limit>-<period>"
//
// Valid periods:
// * "S": second
// * "M": minute
// * "H": hour
//
// Examples:
// * 5 requests / second : "5-S"
// * 5 requests / minute : "5-M"
// * 5 requests / hour : "5-H"
type Activity struct {
	limiter *limiter.Limiter

	sync.RWMutex
	context           map[string]*Context
	filter, threshold float64
}

func (a *Activity) filterRequests(token string) (bool, float64) {
	a.RLock()
	context := a.context[token]
	a.RUnlock()
	if context == nil {
		context = &Context{
			prev: MemorySize - 1,
		}
		a.Lock()
		a.context[token] = context
		a.Unlock()
	}

	context.Lock()
	time := time.Now().UnixNano()
	previous := context.memory[context.prev]
	context.memory[context.index] = time
	context.index, context.prev = (context.index+1)%MemorySize, context.index
	size, valid := context.size, true
	if size < MemorySize {
		size++
		context.size, valid = size, false
	}
	oldest := context.memory[context.index]
	context.Unlock()

	alpha := float64(time-previous) / float64(time-oldest)
	rate := float64(size) / float64(time-oldest)
	a.Lock()
	a.filter = alpha*rate + (1-alpha)*a.filter
	filtered := rate / a.filter
	a.Unlock()

	return valid, filtered
}

// New creates a new rate limiter
func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := Settings{}
	err := metadata.MapToStruct(ctx.Settings(), &settings, true)
	if err != nil {
		return nil, err
	}

	logger := ctx.Logger()
	logger.Debugf("Setting: %b", settings)

	rate, err := limiter.NewRateFromFormatted(settings.Limit)
	if err != nil {
		panic(err)
	}
	store := memory.NewStore()
	limiter := limiter.New(store, rate)

	act := Activity{
		limiter:   limiter,
		context:   make(map[string]*Context, 256),
		threshold: settings.SpikeThreshold,
	}

	return &act, nil
}

// Metadata return the metadata for the activity
func (a *Activity) Metadata() *activity.Metadata {
	return activityMetadata
}

// Eval executes the activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	input := Input{}
	err = ctx.GetInputObject(&input)
	if err != nil {
		return false, err
	}

	output := Output{}

	// check for request token
	if input.Token == "" {
		output.Error = true
		output.ErrorMessage = "Token not found"

		err = ctx.SetOutputObject(&output)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// consume limit
	limiterContext, err := a.limiter.Get(context.TODO(), input.Token)
	if err != nil {
		return true, nil
	}

	var (
		valid    bool
		filtered float64
	)
	if a.threshold != 0 {
		valid, filtered = a.filterRequests(input.Token)
	}

	// check the ratelimit
	output.LimitAvailable = limiterContext.Remaining
	if limiterContext.Reached || (valid && filtered > a.threshold) {
		output.LimitReached = true
	} else {
		output.LimitReached = false
	}

	err = ctx.SetOutputObject(&output)
	if err != nil {
		return false, err
	}

	return true, nil
}
