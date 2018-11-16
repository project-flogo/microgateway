package circuitbreaker

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
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
	"github.com/project-flogo/microgateway/activity/circuitbreaker/example"
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

func TestCircuitBreakerModeA(t *testing.T) {
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		now = time.Now
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
	now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		now = time.Now
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
	now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		now = time.Now
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
	now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		now = time.Now
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
	p := activity.(*Activity).context.Probability(now())
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
	p = activity.(*Activity).context.Probability(now())
	assert.Equal(t, 0.0, math.Floor(p*100))
}

type handler struct {
	Slow bool
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if h.Slow {
		time.Sleep(10 * time.Second)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(reply))
	if err != nil {
		panic(err)
	}
}

const reply = `{
	"id": 1,
	"category": {
		"id": 0,
		"name": "string"
	},
	"name": "sally",
	"photoUrls": ["string"],
	"tags": [{ "id": 0,"name": "string" }],
	"status":"available"
}`

// Response is a reply form the server
type Response struct {
	Pet    json.RawMessage `json:"pet"`
	Status string          `json:"status"`
	Error  string          `json:"error"`
}

func testApplication(t *testing.T, e engine.Engine) {
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		now = time.Now
	}()

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
	_, err := http.Get("http://localhost:1234/pets/1")
	for err != nil {
		_, err = http.Get("http://localhost:1234/pets/1")
	}
	defer s.Shutdown(context.Background())

	err = e.Start()
	assert.Nil(t, err)
	defer func() {
		e.Stop()
	}()

	client := &http.Client{
		Transport: &http.Transport{},
	}

	var r interface{}
	err = json.Unmarshal([]byte(reply), &r)
	assert.Nil(t, err)
	data, err := json.Marshal(r)
	assert.Nil(t, err)

	request := func() Response {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets/1", nil)
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		var rsp Response
		err = json.Unmarshal(body, &rsp)
		assert.Nil(t, err)
		clock = clock.Add(time.Second)
		return rsp
	}
	response := request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))

	s.Shutdown(context.Background())
	for i := 0; i < 5; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}

	response = request()
	assert.Equal(t, "circuit breaker tripped", response.Error)

	sr := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		sr.ListenAndServe()
	}()
	_, err = http.Get("http://localhost:1234/pets/1")
	for err != nil {
		_, err = http.Get("http://localhost:1234/pets/1")
	}
	defer sr.Shutdown(context.Background())

	clock = clock.Add(60 * time.Second)
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
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
