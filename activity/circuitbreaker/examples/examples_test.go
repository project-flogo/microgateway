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

func testApplication(t *testing.T, e engine.Engine) {
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

	s.Shutdown(context.Background())
	transport.CloseIdleConnections()
	for i := 0; i < 5; i++ {
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

	e, err := Example()
	assert.Nil(t, err)
	testApplication(t, e)
}

func TestIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}

	data, err := ioutil.ReadFile(filepath.FromSlash("./json/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testApplication(t, e)
}
