package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/project-flogo/contrib/activity/rest"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway/activity/jwt"
	_ "github.com/project-flogo/contrib/activity/channel"
	"github.com/project-flogo/microgateway/api"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

// Response is a reply form the server
type Response struct {
	Error string `json:"error"`
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

func testBasicGatewayApplication(t *testing.T, e engine.Engine) {
	defer api.ClearResources()
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
	request := func() string {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets/1", nil)
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body := request()
	assert.NotEqual(t, 0, string(body))
}

func TestBasicGatewayIntegrationAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Basic Gateway API integration test in short mode")
	}

	e, err := BasicGatewayExample()
	assert.Nil(t, err)
	testBasicGatewayApplication(t, e)
}

func TestBasicGatewayIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Basic Gateway JSON integration test in short mode")
	}

	data, err := ioutil.ReadFile(filepath.FromSlash("./json/basic-gateway/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testBasicGatewayApplication(t, e)
}

func testHandlerRoutingApplication(t *testing.T, e engine.Engine) {
	defer api.ClearResources()
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
	request := func(auth string, id int) (string, Response) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:9096/pets/%d", id), nil)
		assert.Nil(t, err)
		if auth != "" {
			req.Header.Add("Auth", auth)
		}
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		var rsp Response
		err = json.Unmarshal(body, &rsp)
		assert.Nil(t, err)
		return string(body), rsp
	}

	body, response := request("", 1)
	assert.Equal(t, "", response.Error)
	assert.NotEqual(t, 0, len(body))

	_, response = request("", 8)
	assert.Equal(t, "id must be less than 8", response.Error)

	body, _ = request("1337", 8)
	assert.NotEqual(t, 0, len(body))
}

func TestHandlerRoutingIntegrationAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Handler Routing API integration test in short mode")
	}

	e, err := HandlerRoutingExample()
	assert.Nil(t, err)
	testHandlerRoutingApplication(t, e)
}

func TestHandlerRoutingIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Handler Routing JSON integration test in short mode")
	}

	data, err := ioutil.ReadFile(filepath.FromSlash("./json/handler-routing/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testHandlerRoutingApplication(t, e)
}

func testDefaultHTTPPattern(t *testing.T, e engine.Engine) {
	defer api.ClearResources()
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
	request := func() string {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets/1", nil)
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body := request()
	assert.NotEqual(t, 0, len(body))
}

func TestDefaultHttpPatternAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Basic Gateway API integration test in short mode")
	}

	e, err := DefaultHTTPPattern()
	assert.Nil(t, err)
	testDefaultHTTPPattern(t, e)
}

func TestDefaultHttpPatternJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Basic Gateway JSON integration test in short mode")
	}
	data, err := ioutil.ReadFile(filepath.FromSlash("./json/default-http-pattern/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testDefaultHTTPPattern(t, e)
}


func testDefaultChannelPattern(t *testing.T, e engine.Engine) {
	defer api.ClearResources()

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
	request := func() string {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/endpoint", nil)
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body := request()
	assert.NotEqual(t, 0, len(body))
}

func TestDefaultChannelPatternAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Basic Gateway API integration test in short mode")
	}

	e, err := DefaultChannelPattern()
	assert.Nil(t, err)
	testDefaultChannelPattern(t, e)
}

func TestDefaultChannelPatternJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Basic Gateway JSON integration test in short mode")
	}
	data, err := ioutil.ReadFile(filepath.FromSlash("./json/default-channel-pattern/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testDefaultChannelPattern(t, e)
}
