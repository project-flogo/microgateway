package jwt

import (
	"encoding/json"
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
	"github.com/project-flogo/microgateway/activity/jwt/example"
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

func TestJWT(t *testing.T) {
	activity, err := New(newInitContext(nil))
	assert.Nil(t, err)
	execute := func(serviceName string, values map[string]interface{}, should error) {
		_, err := activity.Eval(newActivityContext(values))
		assert.Equal(t, should, err)
	}

	inputValues := map[string]interface{}{
		"signingMethod": "HMAC",
		"key":           "qwertyuiopasdfghjklzxcvbnm789101",
		"aud":           "www.mashling.io",
		"iss":           "Mashling",
		"sub":           "tempuser@mail.com",
		"token":         "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0MDQ4NzgzNywiZXhwIjoxNTcyMDIzODM4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsImlkIjoiMSJ9.-Tzfn5ZS0kM-u07qkpFrDxdyptBJIvLesuUzVXdqn48",
	}
	execute("reset", inputValues, nil)
}

// Response is a reply form the server
type Response struct {
	Pet   json.RawMessage `json:"pet"`
	Error json.RawMessage `json:"error"`
}

type Error struct {
	ValidationMessage string `json:"validationMessage"`
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

	request := func(auth string) Response {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets", nil)
		assert.Nil(t, err)
		if auth != "" {
			req.Header.Add("Authorization", auth)
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

	response := request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0MDQ4NzgzNywiZXhwIjoxNTcyMDIzODM4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsImlkIjoiMSJ9.-Tzfn5ZS0kM-u07qkpFrDxdyptBJIvLesuUzVXdqn48")
	assert.Equal(t, "\"JWT token is valid\"", string(response.Error))
	assert.Condition(t, func() bool {
		return len(response.Pet) > 0
	})

	response = request("Bearer <Access_Token>")
	var er Error
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "token contains an invalid number of segments", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))

	response = request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0MDQ4NzgzNywiZXhwIjoxNTcyMDIzODM4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsImlkIjoiMSJ9.-Tzfn5ZS0kM-u07qkpFrDxdyptBJIvLesuUzVXdqn4")
	er = Error{}
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "signature is invalid", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))
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
