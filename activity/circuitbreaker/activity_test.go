package circuitbreaker

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	logger "github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/stretchr/testify/assert"
)

type initContext struct {
	settings map[string]interface{}
}

func newInitContext(values map[string]interface{}) *initContext {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &initContext{
		settings: values,
	}
}

func (i *initContext) Settings() map[string]interface{} {
	return i.settings
}

func (i *initContext) MapperFactory() mapper.Factory {
	return nil
}

func (i *initContext) Logger() logger.Logger {
	return logger.RootLogger()
}

type activityContext struct {
	input  map[string]interface{}
	output map[string]interface{}
}

func newActivityContext(values map[string]interface{}) *activityContext {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &activityContext{
		input:  values,
		output: make(map[string]interface{}),
	}
}

func (a *activityContext) ActivityHost() activity.Host {
	return a
}

func (a *activityContext) Name() string {
	return "test"
}

func (a *activityContext) GetInput(name string) interface{} {
	return a.input[name]
}

func (a *activityContext) SetOutput(name string, value interface{}) error {
	a.output[name] = value
	return nil
}

func (a *activityContext) GetInputObject(input data.StructValue) error {
	return input.FromMap(a.input)
}

func (a *activityContext) SetOutputObject(output data.StructValue) error {
	a.output = output.ToMap()
	return nil
}

func (a *activityContext) GetSharedTempData() map[string]interface{} {
	return nil
}

func (a *activityContext) ID() string {
	return "test"
}

func (a *activityContext) IOMetadata() *metadata.IOMetadata {
	return nil
}

func (a *activityContext) Reply(replyData map[string]interface{}, err error) {

}

func (a *activityContext) Return(returnData map[string]interface{}, err error) {

}

func (a *activityContext) Scope() data.Scope {
	return nil
}

func (a *activityContext) Logger() logger.Logger {
	return logger.RootLogger()
}

func (a *activityContext) GetTracingContext() trace.TracingContext {
	return nil
}

func TestCircuitBreakerModeA(t *testing.T) {
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		Now = time.Now
	}()

	activity, err := New(newInitContext(nil))
	assert.Nil(t, err)
	execute := func(serviceName string, values map[string]interface{}, should error) {
		_, err := activity.Eval(newActivityContext(values))
		assert.Equal(t, should, err)
	}

	for i := 0; i < 4; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	execute("reset", nil, nil)
	execute("reset", map[string]interface{}{"operation": "reset"}, nil)

	for i := 0; i < 5; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	execute("reset", nil, ErrorCircuitBreakerTripped)

	clock = clock.Add(60 * time.Second)

	execute("reset", nil, nil)
	execute("reset", map[string]interface{}{"operation": "counter"}, nil)

	execute("reset", nil, ErrorCircuitBreakerTripped)

	clock = clock.Add(60 * time.Second)

	execute("reset", nil, nil)
}

func TestCircuitBreakerModeB(t *testing.T) {
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		Now = time.Now
	}()

	activity, err := New(newInitContext(map[string]interface{}{
		"mode": CircuitBreakerModeB,
	}))
	assert.Nil(t, err)
	execute := func(serviceName string, values map[string]interface{}, should error) {
		_, err := activity.Eval(newActivityContext(values))
		assert.Equal(t, should, err)
	}

	for i := 0; i < 4; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	clock = clock.Add(60 * time.Second)

	for i := 0; i < 5; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	execute("reset", nil, ErrorCircuitBreakerTripped)

	clock = clock.Add(60 * time.Second)

	execute("reset", nil, nil)
	execute("reset", map[string]interface{}{"operation": "counter"}, nil)

	execute("reset", nil, ErrorCircuitBreakerTripped)

	clock = clock.Add(60 * time.Second)

	execute("reset", nil, nil)
}

func TestCircuitBreakerModeC(t *testing.T) {
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		Now = time.Now
	}()

	activity, err := New(newInitContext(map[string]interface{}{
		"mode": CircuitBreakerModeC,
	}))
	assert.Nil(t, err)
	execute := func(serviceName string, values map[string]interface{}, should error) {
		_, err := activity.Eval(newActivityContext(values))
		assert.Equal(t, should, err)
	}

	for i := 0; i < 4; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	clock = clock.Add(60 * time.Second)

	for i := 0; i < 4; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	execute("reset", nil, nil)
	execute("reset", map[string]interface{}{"operation": "reset"}, nil)

	for i := 0; i < 5; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "counter"}, nil)
	}

	execute("reset", nil, ErrorCircuitBreakerTripped)

	clock = clock.Add(60 * time.Second)

	execute("reset", nil, nil)
	execute("reset", map[string]interface{}{"operation": "counter"}, nil)

	execute("reset", nil, ErrorCircuitBreakerTripped)

	clock = clock.Add(60 * time.Second)

	execute("reset", nil, nil)
}

func TestCircuitBreakerModeD(t *testing.T) {
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		Now = time.Now
	}()

	activity, err := New(newInitContext(map[string]interface{}{
		"mode": CircuitBreakerModeD,
	}))
	assert.Nil(t, err)
	execute := func(serviceName string, values map[string]interface{}, should error) error {
		_, err := activity.Eval(newActivityContext(values))
		assert.Equal(t, should, err)
		return err
	}

	for i := 0; i < 100; i++ {
		execute("reset", nil, nil)
		execute("reset", map[string]interface{}{"operation": "reset"}, nil)
	}

	p := activity.(*Activity).context.Probability(Now())
	assert.Equal(t, 0.0, math.Floor(p*100))

	type Test struct {
		a, b error
	}
	tests := []Test{
		{nil, nil},
		{nil, nil},
		{ErrorCircuitBreakerTripped, nil},
		{ErrorCircuitBreakerTripped, nil},
		{nil, nil},
		{ErrorCircuitBreakerTripped, nil},
		{ErrorCircuitBreakerTripped, nil},
		{ErrorCircuitBreakerTripped, nil},
	}
	for _, test := range tests {
		err := execute("reset", nil, test.a)
		if err != nil {
			continue
		}
		execute("reset", map[string]interface{}{"operation": "counter"}, test.b)
	}

	tests = []Test{
		{nil, nil},
		{nil, nil},
		{nil, nil},
		{nil, nil},
		{nil, nil},
	}
	for _, test := range tests {
		err := execute("reset", nil, test.a)
		if err != nil {
			continue
		}
		execute("reset", map[string]interface{}{"operation": "reset"}, test.b)
	}
	p = activity.(*Activity).context.Probability(Now())
	assert.Equal(t, 0.0, math.Floor(p*100))
}
