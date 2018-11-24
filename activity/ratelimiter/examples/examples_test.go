package examples

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/project-flogo/core/engine"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

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
	test.Pour("9096")

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
