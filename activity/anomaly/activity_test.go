package anomaly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/project-flogo/contrib/activity/rest"
	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	logger "github.com/project-flogo/core/support/log"
	"github.com/project-flogo/microgateway"
	microapi "github.com/project-flogo/microgateway/api"
	"github.com/stretchr/testify/assert"
)

var complexityTests = []string{`{
 "alfa": [
  {"alfa": "1"},
	{"bravo": "2"}
 ],
 "bravo": [
  {"alfa": "3"},
	{"bravo": "4"}
 ]
}`, `{
 "a": [
  {"a": "aa"},
	{"b": "bb"}
 ],
 "b": [
  {"a": "aa"},
	{"b": "bb"}
 ]
}`}

func generateRandomJSON(rnd *rand.Rand) map[string]interface{} {
	sample := func(stddev float64) int {
		return int(math.Abs(rnd.NormFloat64()) * stddev)
	}
	sampleCount := func() int {
		return sample(1) + 1
	}
	sampleName := func() string {
		const symbols = 'z' - 'a'
		s := sample(8)
		if s > symbols {
			s = symbols
		}
		return string('a' + s)
	}
	sampleValue := func() string {
		value := sampleName()
		return value + value
	}
	sampleDepth := func() int {
		return sample(3)
	}
	var generate func(hash map[string]interface{}, depth int)
	generate = func(hash map[string]interface{}, depth int) {
		count := sampleCount()
		if depth > sampleDepth() {
			for i := 0; i < count; i++ {
				hash[sampleName()] = sampleValue()
			}
			return
		}
		for i := 0; i < count; i++ {
			array := make([]interface{}, sampleCount())
			for j := range array {
				sub := make(map[string]interface{})
				generate(sub, depth+1)
				array[j] = sub
			}
			hash[sampleName()] = array
		}
	}
	object := make(map[string]interface{})
	generate(object, 0)
	return object
}

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

func TestActivity(t *testing.T) {
	activity, err := New(newInitContext(nil))
	assert.Nil(t, err)

	eval := func(data []byte) float32 {
		var payload interface{}
		err := json.Unmarshal(data, &payload)
		assert.Nil(t, err)
		ctx := newActivityContext(map[string]interface{}{"payload": payload})
		_, err = activity.Eval(ctx)
		assert.Nil(t, err)
		return ctx.output["complexity"].(float32)
	}

	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < 1024; i++ {
		data, err := json.Marshal(generateRandomJSON(rnd))
		assert.Nil(t, err)
		eval(data)
	}
	a := eval([]byte(complexityTests[0]))
	b := eval([]byte(complexityTests[1]))
	assert.Condition(t, func() (success bool) {
		return a > b
	}, "complexity sanity check failed")
}

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

// Response is a reply form the server
type Response struct {
	Error      string  `json:"error"`
	Complexity float64 `json:"complexity"`
}

var anomalyPayload = `{
 "alfa": [
  {"alfa": "1"},
	{"bravo": "2"}
 ],
 "bravo": [
  {"alfa": "3"},
	{"bravo": "4"}
 ]
}`

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testHandler := handler{}
	s := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		s.ListenAndServe()
	}()
	_, err := http.Get("http://localhost:1234/test")
	for err != nil {
		_, err = http.Get("http://localhost:1234/test")
	}
	defer s.Shutdown(context.Background())

	app := api.NewApp()

	gateway := microapi.New("Test")
	serviceAnomaly := gateway.NewService("Anomaly", &Activity{})
	serviceAnomaly.SetDescription("Look for anomalies")
	serviceAnomaly.AddSetting("depth", 3)

	serviceUpdate := gateway.NewService("Update", &rest.Activity{})
	serviceUpdate.SetDescription("Make calls to test")
	serviceUpdate.AddSetting("uri", "http://localhost:1234/test")
	serviceUpdate.AddSetting("method", "PUT")

	step := gateway.NewStep(serviceAnomaly)
	step.AddInput("payload", "=$.payload.content")
	step = gateway.NewStep(serviceUpdate)
	step.SetIf("($.Anomaly.outputs.count < 100) || ($.Anomaly.outputs.complexity < 3)")
	step.AddInput("content", "=$.payload.content")

	response := gateway.NewResponse(false)
	response.SetIf("($.Anomaly.outputs.count < 100) || ($.Anomaly.outputs.complexity < 3)")
	response.SetCode(200)
	response.SetData("=$.Update.outputs.data")
	response = gateway.NewResponse(true)
	response.SetCode(403)
	response.SetData(map[string]interface{}{
		"error":      "anomaly!",
		"complexity": "=$.Anomaly.outputs.complexity",
	})

	settings, err := gateway.AddResource(app)
	assert.Nil(t, err)

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "PUT",
		Path:   "/test",
	})
	assert.Nil(t, err)

	_, err = handler.NewAction(&microgateway.Action{}, settings)
	assert.Nil(t, err)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()
	defer e.Stop()

	rnd, client := rand.New(rand.NewSource(1)), &http.Client{}
	for i := 0; i < 1024; i++ {
		data, err := json.Marshal(generateRandomJSON(rnd))
		assert.Nil(t, err)
		req, err := http.NewRequest(http.MethodPut, "http://localhost:9096/test", bytes.NewReader(data))
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
		assert.NotEqual(t, "anomaly!", rsp.Error)
	}

	{
		req, err := http.NewRequest(http.MethodPut, "http://localhost:9096/test", bytes.NewBufferString(anomalyPayload))
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
		assert.Equal(t, "anomaly!", rsp.Error)
		fmt.Println(rsp.Complexity)
		assert.Condition(t, func() bool {
			return rsp.Complexity > 8
		})
	}
}
