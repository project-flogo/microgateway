package examples

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/api"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

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
