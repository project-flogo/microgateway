package obfusacte

import (
	"bytes"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

func init() {
	_ = activity.Register(&Activity{}, New) //activity.Register(&Activity{}, New) to create instances using factory method 'New'
}

// Function which defines how to obfuscate the value
type obfuscateFunc func(a string) string

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

//New optional factory method, should be used if one activity instance per configuration is desired
func New(ctx activity.InitContext) (activity.Activity, error) {

	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	var op obfuscateFunc

	// Set the operation.
	if s.Operation == "setLastFour" {
		op = setLastFour
	}

	act := &Activity{settings: s, operation: op}

	return act, nil
}

type Activity struct {
	settings  *Settings
	operation obfuscateFunc
}

func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}
	payload := input.Payload

	// Iterate over the keys for which the obfuscate function should apply.
	for _, val := range a.settings.Fields {
		payload = obfuscate(a.operation, val.(string), payload)
	}

	ctx.SetOutput("result", payload)

	return true, nil
}

// Onfuscate takes in the obfuscate function, key and the payload
// and returns the string where the value of the key is obfuscated.
func obfuscate(op obfuscateFunc, key, payload string) string {
	// Get the index where the key ends.
	// Eg "key":"13445". Should return 6.
	keyEndIndex := strings.Index(payload, key) + len(key) + 3

	//Get the index where the value corresponding to that key ends.
	//Eg. with the above eg it should return 12
	valEndIndex := keyEndIndex + strings.Index(payload[keyEndIndex+1:], "\"")

	//Get the value corresponding to the key.
	keyVal := payload[keyEndIndex:valEndIndex]

	// Apply obfuscate function.
	val := op(keyVal)

	// Stich the result
	result := payload[:keyEndIndex] + val + payload[valEndIndex:]

	return result
}

func setLastFour(val string) string {

	var buffer bytes.Buffer

	for key, v := range val {
		if key < len(val)-3 {
			buffer.WriteString("*")
		} else {
			buffer.WriteString(string(v))
		}

	}

	return buffer.String()

}
