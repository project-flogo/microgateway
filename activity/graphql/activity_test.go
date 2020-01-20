package graphql

import (
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

func TestGraphQLModeA(t *testing.T) {
	activity, err := New(newInitContext(map[string]interface{}{
		"mode": "a",
	}))
	assert.Nil(t, err)

	// schemaFile value not set
	query := `{"query":"query {stationWithEvaId(evaId: 8000105) { name } }"}`
	ctx := newActivityContext(map[string]interface{}{
		"query":         query,
		"maxQueryDepth": 2,
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["error"].(bool))
	assert.Equal(t, ctx.output["errorMessage"].(string), "Schema file is required")

	// schemaFile value not available
	ctx = newActivityContext(map[string]interface{}{
		"query":         query,
		"schemaFile":    "notafile.graphql",
		"maxQueryDepth": 2,
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["error"].(bool))
	assert.Contains(t, ctx.output["errorMessage"].(string), "Not able to read the schema file")

	// valid query
	ctx = newActivityContext(map[string]interface{}{
		"query":         query,
		"schemaFile":    "examples/schema.graphql",
		"maxQueryDepth": 2,
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["valid"].(bool))

	invalidquery := `{"query":"query {stationWithEvaId(evaId: 8000105) { cityname } }"}`
	ctx = newActivityContext(map[string]interface{}{
		"query":         invalidquery,
		"schemaFile":    "examples/schema.graphql",
		"maxQueryDepth": 2,
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["error"].(bool))

	querydepth := `{"query":"{stationWithEvaId(evaId: 8000105) {name location { latitude longitude } picture { url } } }"}`
	ctx = newActivityContext(map[string]interface{}{
		"query":         querydepth,
		"schemaFile":    "examples/schema.graphql",
		"maxQueryDepth": 2,
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["error"].(bool))
}

func TestGraphQLModeB(t *testing.T) {
	activity, err := New(newInitContext(map[string]interface{}{
		"mode":  "b",
		"limit": "500-100-1000",
	}))
	assert.Nil(t, err)

	query := `{"query":"query {stationWithEvaId(evaId: 8000105) { name } }"}`
	ctx := newActivityContext(map[string]interface{}{
		"token":     "token123",
		"operation": "startconsume",
		"query":     query,
	})
	_, err = activity.Eval(ctx)
	assert.False(t, ctx.output["error"].(bool))

	ctx = newActivityContext(map[string]interface{}{
		"token":     "token123",
		"operation": "stopconsume",
	})
	status, err := activity.Eval(ctx)

	assert.True(t, status)
	status, err = activity.Eval(ctx)

	// time.Sleep(900 * time.Millisecond)
	ctx = newActivityContext(map[string]interface{}{
		"token":     "token123",
		"operation": "startconsume",
		"query":     query,
	})
	_, err = activity.Eval(ctx)
	assert.False(t, ctx.output["error"].(bool))
	time.Sleep(400 * time.Millisecond)
	ctx = newActivityContext(map[string]interface{}{
		"token":     "token123",
		"operation": "stopconsume",
	})
	status, err = activity.Eval(ctx)
	assert.True(t, status)

	ctx = newActivityContext(map[string]interface{}{
		"token":     "token123",
		"operation": "startconsume",
		"query":     query,
	})
	_, err = activity.Eval(ctx)
	assert.False(t, ctx.output["error"].(bool))
	time.Sleep(200 * time.Millisecond)
	ctx = newActivityContext(map[string]interface{}{
		"token":     "token123",
		"operation": "stopconsume",
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["error"].(bool))

	ctx = newActivityContext(map[string]interface{}{
		"token":     "token1234",
		"operation": "startconsume",
		"query":     query,
	})
	_, err = activity.Eval(ctx)
	assert.False(t, ctx.output["error"].(bool))
	time.Sleep(100 * time.Millisecond)
	ctx = newActivityContext(map[string]interface{}{
		"token":     "token1234",
		"operation": "stopconsume",
	})
	status2, err := activity.Eval(ctx)
	// status, err = activity.Eval(ctx)
	assert.True(t, status2)

	ctx = newActivityContext(map[string]interface{}{
		"token":     "token1234",
		"operation": "startconsume",
		"query":     query,
	})
	_, err = activity.Eval(ctx)
	assert.False(t, ctx.output["error"].(bool))
	time.Sleep(401 * time.Millisecond)
	ctx = newActivityContext(map[string]interface{}{
		"token":     "token1234",
		"operation": "stopconsume",
	})
	_, err = activity.Eval(ctx)
	assert.True(t, ctx.output["error"].(bool))
}
