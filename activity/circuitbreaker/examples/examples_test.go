package examples

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"testing"
	"time"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/circuitbreaker"
	"github.com/project-flogo/microgateway/api"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

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


func testApplicationModeA(t *testing.T, e engine.Engine, threshold int, timeout int, period int) {
	if threshold == 0{
		threshold = 5
	}
	limit := threshold
	if timeout == 0{
		timeout = 60
	}
	if period == 0{
		period = 60
	}
	defer api.ClearResources()
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	circuitbreaker.Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		circuitbreaker.Now = time.Now
	}()

	test.Drain("1234")
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
	test.Pour("1234")
	defer s.Shutdown(context.Background())

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

	//testcase_1: circuit trips for continous errors
	s.Shutdown(context.Background())
	transport.CloseIdleConnections()

	for i := 0; i < threshold; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}
	response = request()
	assert.Equal(t, "circuit breaker tripped", response.Error)

	test.Drain("1234")
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
	test.Pour("1234")

	clock = clock.Add(time.Duration(timeout) * time.Second)
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))

	//testcase_2: circuit does not trip for non-continous errors
	sr.Shutdown(context.Background())
	transport.CloseIdleConnections()
	threshold = limit
	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}
	test.Drain("1234")
	srv := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		srv.ListenAndServe()
	}()
	test.Pour("1234")
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
	srv.Shutdown(context.Background())
	transport.CloseIdleConnections()
	threshold = limit

	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}
}

func testApplicationModeB(t *testing.T, e engine.Engine, threshold int, timeout int, period int) {
	if threshold == 0{
		threshold = 5
	}
	limit := threshold
	if timeout == 0{
		timeout = 60
	}
	if period == 0{
		period = 60
	}
	defer api.ClearResources()
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	circuitbreaker.Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		circuitbreaker.Now = time.Now
	}()

	test.Drain("1234")
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
	test.Pour("1234")
	defer s.Shutdown(context.Background())

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

	//testcase_1: circuit trips for errors within given period
	s.Shutdown(context.Background())
	transport.CloseIdleConnections()

	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}

	test.Drain("1234")
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
	test.Pour("1234")

	//valid call within period time
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
	sr.Shutdown(context.Background())
	transport.CloseIdleConnections()


	//threshold value failure
	response = request()
	assert.Equal(t, "connection failure", response.Error)

	//trip condition
	response = request()
	assert.Equal(t, "circuit breaker tripped", response.Error)

	test.Drain("1234")
	srv := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		srv.ListenAndServe()
	}()
	test.Pour("1234")

	clock = clock.Add(time.Duration(timeout) * time.Second)
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))

	//testcase_2: Circuit does not trip if number of calls within period does not exceed threshold
	clock = clock.Add(time.Duration(period) * time.Second)
	srv.Shutdown(context.Background())
	transport.CloseIdleConnections()
	threshold = limit
	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}

	clock = clock.Add(time.Duration(period) * time.Second)
	threshold = limit

	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}
	test.Drain("1234")
	srv2 := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		srv2.ListenAndServe()
	}()
	test.Pour("1234")
	defer srv2.Shutdown(context.Background())
	//valid call within period time
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
}

func testApplicationModeC(t *testing.T, e engine.Engine, threshold int, timeout int, period int) {
	if threshold == 0{
		threshold = 5
	}
	limit := threshold
	if timeout == 0{
		timeout = 60
	}
	if period == 0{
		period = 60
	}
	defer api.ClearResources()
	rand.Seed(1)
	clock := time.Unix(1533930608, 0)
	circuitbreaker.Now = func() time.Time {
		now := clock
		clock = clock.Add(time.Duration(rand.Intn(2)+1) * time.Second)
		return now
	}
	defer func() {
		circuitbreaker.Now = time.Now
	}()

	test.Drain("1234")
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
	test.Pour("1234")
	defer s.Shutdown(context.Background())

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

	//testcase_1: circuit trips for continous error within period
	s.Shutdown(context.Background())
	transport.CloseIdleConnections()

	for i := 0; i < threshold; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}

	response = request()
	assert.Equal(t, "circuit breaker tripped", response.Error)

	clock = clock.Add(time.Duration(timeout) * time.Second)

	//testcase_2: circuit does not trip for non-continous errors
	test.Drain("1234")
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
	test.Pour("1234")

	//valid call within period time
	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
	sr.Shutdown(context.Background())
	transport.CloseIdleConnections()

	threshold = limit
	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}

	test.Drain("1234")
	srv := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		srv.ListenAndServe()
	}()
	test.Pour("1234")

	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
	srv.Shutdown(context.Background())
	transport.CloseIdleConnections()

	threshold = limit
	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}


	//testcase_3: circuit does not trip for continous errors which does not occur within give period
	test.Drain("1234")
	srv2 := &http.Server{
		Addr:           ":1234",
		Handler:        &testHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		srv2.ListenAndServe()
	}()
	test.Pour("1234")

	response = request()
	assert.Equal(t, "available", response.Status)
	assert.Equal(t, len(data), len(response.Pet))
	srv2.Shutdown(context.Background())
	transport.CloseIdleConnections()

	threshold = limit
	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}

	clock = clock.Add(time.Duration(period) * time.Second)

	threshold = limit
	for i := 0; i < threshold-1; i++ {
		response := request()
		assert.Equal(t, "connection failure", response.Error)
	}
}


func TestIntegrationAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping API integration test in short mode")
	}
	parameters := []struct {
		mode string
		threshold int //0 is default setting
		timeout int //0 is default setting
		period int //for mode b and c
	}{
		//{"mode","threshold","timeout","period"}
		{"a",2,0,0},{"a",0,0,0},{"a",4,30,0},
		{"b",4,30,20},{"b",0,30,20},{"b",0,0,20},{"b",4,30,20},{"b",0,0,20},
		{"c",4,30,20},{"c",0,30,20},{"c",0,0,20},{"c",4,30,20},{"c",0,0,20},
	}
	for i := range parameters {
		e, err := Example(parameters[i].mode,parameters[i].threshold,parameters[i].timeout,parameters[i].period)
		assert.Nil(t, err)
		if parameters[i].mode == "a"{
			testApplicationModeA(t, e,parameters[i].threshold,parameters[i].timeout,parameters[i].period)
		}else if parameters[i].mode == "b"{
			testApplicationModeB(t, e,parameters[i].threshold,parameters[i].timeout, parameters[i].period)
		}else if parameters[i].mode == "c"{
			testApplicationModeC(t, e,parameters[i].threshold,parameters[i].timeout, parameters[i].period)
		}
	}
}


func TestIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}
	parameters := []struct {
		mode string
		threshold int //0 is default setting
		timeout int //0 is default setting
		period int //or mode b and c
	}{
		//{"mode","threshold","timeout","period"}
		{"a",2,0,0},{"a",0,0,0},{"a",4,30,0},
		{"b",4,30,20},{"b",0,30,20},{"b",0,0,20},{"b",4,30,20},{"b",0,0,20},
		{"c",4,30,20},{"c",0,30,20},{"c",0,0,20},{"c",4,30,20},{"c",0,0,20},
	}
	data, err := ioutil.ReadFile(filepath.FromSlash("./json/flogo.json"))
	assert.Nil(t, err)
	for i := range parameters {
		var Input input
		err = json.Unmarshal(data, &Input)
		if parameters[i].threshold == 0{
			parameters[i].threshold = 5
		}
		if parameters[i].timeout == 0{
			parameters[i].timeout = 60
		}
		if parameters[i].period == 0{
			parameters[i].period = 60
		}
		Input.Resources[0].Data.Services[0]["settings"] = map[string]interface{}{"mode":parameters[i].mode,
			"threshold":parameters[i].threshold, "timeout":parameters[i].timeout, "period":parameters[i].period}
		data, _ = json.Marshal(Input)
		cfg, err := engine.LoadAppConfig(string(data), false)
		assert.Nil(t, err)
		e, err := engine.New(cfg)
		assert.Nil(t, err)
		if parameters[i].mode == "a"{
			testApplicationModeA(t, e, parameters[i].threshold, parameters[i].timeout, parameters[i].period)
		}else if parameters[i].mode == "b"{
			testApplicationModeB(t, e, parameters[i].threshold, parameters[i].timeout, parameters[i].period)
		}else if parameters[i].mode == "c"{
			testApplicationModeC(t, e, parameters[i].threshold, parameters[i].timeout, parameters[i].period)
		}
	}
}


//--------data structure-------//

type input struct{
	Name string `json:"name"`
	Type string `json:"type"`
	Version string `json:"version"`
	Desc string `json:"description"`
	Prop interface{} `json:"properties"`
	Channels interface{} `json:"channels"`
	Trig interface{} `json:"triggers"`
	Resources []struct{
		Id string `json:"id"`
		Compress bool `json:"compressed"`
		Data struct{
			   Name string `json:"name"`
			   Steps []interface{} `json:"steps"`
			   Responses []interface{} `json:"responses"`
			   Services []map[string]interface{} `json:"services"`
		   } `json:"data"`
	} `json:"resources"`
	Actions interface{} `json:"actions"`
}
