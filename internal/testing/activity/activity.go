package activity

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

func init() {
	activity.Register(&Activity{}, New)
}

var (
	// Messages is the message the dummy activity got in its input
	Message = ""
	// HasEvaled is true when the dummy activity has been evaluated
	HasEvaled = false
)

var activityMetadata = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// New creates a new dummy activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	act := &Activity{}

	return act, nil
}

// Activity is a dummy activity for testing
type Activity struct {
}

// Metadata returns the metadata for the dummy activity
func (a *Activity) Metadata() *activity.Metadata {
	return activityMetadata
}

// Eval evaluates the dummy activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}

	Message, HasEvaled = input.Message, true

	output := &Output{Data: "1337"}
	err = ctx.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Reset resets the activity for another test
func Reset() {
	Message = ""
	HasEvaled = false
}
