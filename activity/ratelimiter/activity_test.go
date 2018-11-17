package ratelimiter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine"
	logger "github.com/project-flogo/core/support/log"
	"github.com/project-flogo/microgateway/activity/ratelimiter/example"
	test "github.com/project-flogo/microgateway/internal/testing"
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

func TestRatelimiter(t *testing.T) {
	activity, err := New(newInitContext(map[string]interface{}{
		"limit": "1-S",
	}))
	assert.Nil(t, err)

	ctx := newActivityContext(map[string]interface{}{
		"token": "abc123",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.False(t, ctx.output["limitReached"].(bool), "limit should not be reached")

	ctx = newActivityContext(map[string]interface{}{
		"token": "abc123",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.True(t, ctx.output["limitReached"].(bool), "limit should be reached")

	ctx = newActivityContext(map[string]interface{}{
		"token": "sally",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.False(t, ctx.output["limitReached"].(bool), "limit should not be reached")

	time.Sleep(time.Second)

	ctx = newActivityContext(map[string]interface{}{
		"token": "abc123",
	})
	_, err = activity.Eval(ctx)
	assert.Nil(t, err)
	assert.False(t, ctx.output["limitReached"].(bool), "limit should not be reached")
}

// Response is a reply form the server
type Response struct {
	Status string `json:"status"`
}

func testApplication(t *testing.T, e engine.Engine) {
	test.Drain("9096")
	err := e.Start()
	assert.Nil(t, err)
	defer func() {
		err := e.Stop()
		assert.Nil(t, err)
	}()

	transport := &http.Transport{
		MaxIdleConns: 1,
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport: transport,
	}

	request := func(token string) Response {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets/1", nil)
		assert.Nil(t, err)
		if token != "" {
			req.Header.Add("Token", token)
		}
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		var rsp Response
		err = json.Unmarshal(body, &rsp)
		assert.Nil(t, err)
		return rsp
	}

	for i := 0; i < 3; i++ {
		response := request("TOKEN1")
		assert.NotEqual(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)
		assert.NotEqual(t, "Token not found", response.Status)
	}
	response := request("TOKEN1")
	assert.Equal(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)

	response = request("TOKEN2")
	assert.NotEqual(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)
	assert.NotEqual(t, "Token not found", response.Status)

	time.Sleep(time.Minute + 3*time.Second)

	response = request("TOKEN1")
	assert.NotEqual(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)
	assert.NotEqual(t, "Token not found", response.Status)

	response = request("")
	assert.Equal(t, "Token not found", response.Status)
}

func TestIntegrationAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping API integration test in short mode")
	}

	e, err := example.Example(&Activity{})
	assert.Nil(t, err)
	testApplication(t, e)
}

func TestIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}

	data, err := ioutil.ReadFile(filepath.FromSlash("./examples/json/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testApplication(t, e)
}
