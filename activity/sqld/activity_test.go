package sqld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine"
	logger "github.com/project-flogo/core/support/log"
	"github.com/project-flogo/microgateway/activity/sqld/example"
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

func TestSQLD(t *testing.T) {
	activity, err := New(newInitContext(nil))
	assert.Nil(t, err)

	test := func(a string, attack bool) {
		var payload interface{} = map[string]interface{}{
			"content": map[string]interface{}{
				"test": a,
			},
		}
		ctx := newActivityContext(map[string]interface{}{"payload": payload})
		_, err = activity.Eval(ctx)
		assert.Nil(t, err)

		value, attackValues := ctx.output["attack"].(float32), ctx.output["attackValues"].(map[string]interface{})
		if attack {
			assert.Condition(t, func() (success bool) {
				return value > 50
			}, fmt.Sprint("should be an attack", a, value))
			assert.Condition(t, func() (success bool) {
				return attackValues["content"].(map[string]interface{})["test"].(float64) > 50
			}, fmt.Sprint("should be an attack", a, value))
		} else {
			assert.Condition(t, func() (success bool) {
				return value < 50
			}, fmt.Sprint("should not be an attack", a, value))
			assert.Condition(t, func() (success bool) {
				return attackValues["content"].(map[string]interface{})["test"].(float64) < 50
			}, fmt.Sprint("should not be an attack", a, value))
		}
	}
	test("test or 1337=1337 --\"", true)
	test(" or 1=1 ", true)
	test("/**/or/**/1337=1337", true)
	test("abc123", false)
	test("abc123 123abc", false)
	test("123", false)
	test("abcorabc", false)
	test("available", false)
	test("orcat1", false)
	test("cat1or", false)
	test("cat1orcat1", false)
}

// Response is a reply form the server
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

var (
	payload = `{
	  "id": 1,
	  "category": {
	    "id": 0,
	    "name": "string"
	  },
	  "name": "cat",
	  "photoUrls": [
	    "string"
	  ],
	  "tags": [
	    {
	      "id": 0,
	      "name": "string"
	    }
	  ],
	  "status": "available"
	}`
	attackPayload = `{
	  "id": 1,
	  "category": {
	    "id": 0,
	    "name": "string"
	  },
	  "name": " or 1=1 ",
	  "photoUrls": [
	    "string"
	  ],
	  "tags": [
	    {
	      "id": 0,
	      "name": "string"
	    }
	  ],
	  "status": "available"
	}`
)

func testApplication(t *testing.T, e engine.Engine) {
	test.Drain("9096")
	err := e.Start()
	assert.Nil(t, err)
	defer func() {
		err := e.Stop()
		assert.Nil(t, err)
	}()
	test.Pour("9096")

	transport := &http.Transport{
		MaxIdleConns: 1,
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport: transport,
	}
	request := func(payload string) Response {
		req, err := http.NewRequest(http.MethodPut, "http://localhost:9096/pets", bytes.NewReader([]byte(payload)))
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		rsp := Response{}
		err = json.Unmarshal(body, &rsp)
		assert.Nil(t, err)
		return rsp
	}

	response := request(payload)
	assert.NotEqual(t, "", response.Status)

	response = request(attackPayload)
	assert.Equal(t, "hack attack!", response.Error)
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
