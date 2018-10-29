package ratelimiter

import (
	"context"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
)

var (
	activityMetadata = activity.ToMetadata(&Settings{}, &Input{}, &Output{})
)

func init() {
	activity.Register(&Activity{}, New)
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
}

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
		limiter: limiter,
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

	// check the ratelimit
	output.LimitAvailable = limiterContext.Remaining
	if limiterContext.Reached {
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
