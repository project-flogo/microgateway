package sqld

import (
	"os"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/logger"
	"github.com/project-flogo/microgateway/activity/sqld/injectsec"
	"github.com/project-flogo/microgateway/activity/sqld/injectsec/gru"
)

var (
	maker            *injectsec.DetectorMaker
	log              = logger.GetLogger("activity-circuitbreaker")
	activityMetadata = activity.ToMetadata(&Settings{}, &Input{}, &Output{})
)

func init() {
	var err error
	maker, err = injectsec.NewDetectorMaker()
	if err != nil {
		panic(err)
	}

	activity.Register(&Activity{}, New)
}

// Activity is a SQL injection attack detector
type Activity struct {
	Maker *injectsec.DetectorMaker
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := Settings{}
	err := metadata.MapToStruct(ctx.Settings(), &settings, true)
	if err != nil {
		return nil, err
	}

	log.Debugf("Setting: %b", settings)

	act := Activity{}

	if settings.File != "" {
		var in *os.File
		in, err = os.Open(settings.File)
		if err != nil {
			return nil, err
		}
		defer in.Close()
		act.Maker, err = injectsec.NewDetectorMakerWithWeights(in)
		if err != nil {
			return nil, err
		}
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

	var detector *gru.Detector
	if a.Maker != nil {
		detector = a.Maker.Make()
	} else {
		detector = maker.Make()
	}

	output := Output{
		AttackValues: make(map[string]interface{}),
	}

	var testMap func(a, values map[string]interface{}) (err error)
	testMap = func(a, values map[string]interface{}) (err error) {
		for k, v := range a {
			switch element := v.(type) {
			case []interface{}:
				valuesList := make([]interface{}, 0, len(element))
				for _, item := range element {
					switch element := item.(type) {
					case map[string]interface{}:
						childValues := make(map[string]interface{})
						err = testMap(element, childValues)
						if err != nil {
							return
						}
						valuesList = append(valuesList, childValues)
					case string:
						probability, err := detector.Detect(element)
						valuesList = append(valuesList, float64(probability))
						if probability > output.Attack {
							output.Attack = probability
						}
						if err != nil {
							return err
						}
					}
				}
				values[k] = valuesList
			case map[string]interface{}:
				childValues := make(map[string]interface{})
				err = testMap(element, childValues)
				if err != nil {
					return
				}
				values[k] = childValues
			case string:
				probability, err := detector.Detect(element)
				values[k] = float64(probability)
				if probability > output.Attack {
					output.Attack = probability
				}
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	test := func(key string) (err error) {
		if a, ok := input.Payload[key]; ok {
			switch b := a.(type) {
			case []interface{}:
				valuesList := make([]interface{}, 0, len(b))
				for _, item := range b {
					switch element := item.(type) {
					case map[string]interface{}:
						childValues := make(map[string]interface{})
						err = testMap(element, childValues)
						if err != nil {
							return
						}
						valuesList = append(valuesList, childValues)
					case string:
						probability, err := detector.Detect(element)
						valuesList = append(valuesList, float64(probability))
						if probability > output.Attack {
							output.Attack = probability
						}
						if err != nil {
							return err
						}
					}
				}
				output.AttackValues[key] = valuesList
			case map[string]interface{}:
				values := make(map[string]interface{})
				err = testMap(b, values)
				output.AttackValues[key] = values
			case map[string]string:
				values := make(map[string]interface{})
				for _, v := range b {
					probability, err := detector.Detect(v)
					values[v] = float64(probability)
					if probability > output.Attack {
						output.Attack = probability
					}
					if err != nil {
						return err
					}
				}
				output.AttackValues[key] = values
			}
		}

		return
	}

	err = test("pathParams")
	if err != nil {
		return false, err
	}
	err = test("queryParams")
	if err != nil {
		return false, err
	}
	err = test("content")
	if err != nil {
		return false, err
	}

	err = ctx.SetOutputObject(&output)
	if err != nil {
		return false, err
	}

	return true, nil
}
