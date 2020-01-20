package examples

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	_ "github.com/project-flogo/contrib/activity/rest"
	_ "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway"
	_ "github.com/project-flogo/microgateway/activity/graphql"
	"github.com/project-flogo/microgateway/api"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

var url = "http://localhost:9096/graphql"

var validPayload = map[string]string{
	"query": "query {stationWithEvaId(evaId: 8000105){name}}",
}

var invalidPayload = map[string]string{
	"query": "query {stationWithEvaId(evaId: 8000105) { cityname } }",
}

var invalidPayload1 = map[string]string{
	"query": "{stationWithEvaId(evaId: 8000105) {name location { latitude longitude } picture { url } } }",
}

func TestGqlJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}

	data, err := ioutil.ReadFile(filepath.FromSlash("./json/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)

	defer api.ClearResources()
	test.Drain("9096")
	err = e.Start()
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

	payload, _ := json.Marshal(&validPayload)

	request := func() string {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body := request()
	resp := "{\"response\":\"{\\\"data\\\":{\\\"stationWithEvaId\\\":{\\\"name\\\":\\\"Frankfurt (Main) Hbf\\\"}}}\",\"validationMessage\":\"Valid graphQL query. query = {\\\"query\\\":\\\"query {stationWithEvaId(evaId: 8000105){name}}\\\"}\\n type = Query \\n queryDepth = 2\"}\n"
	assert.Equal(t, resp, body)

	payload1, _ := json.Marshal(&invalidPayload)
	request1 := func() string {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload1))
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body1 := request1()
	resp1 := "{\"error\":\"Not a valid graphQL request. Details: [graphql: Cannot query field \\\"cityname\\\" on type \\\"Station\\\". (line 1, column 43)]\"}\n"

	assert.Equal(t, resp1, body1)

	payload2, _ := json.Marshal(&invalidPayload1)
	request2 := func() string {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload2))
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body2 := request2()
	resp2 := "{\"error\":\"graphQL request query depth[3] is exceeded allowed maxQueryDepth[2]\"}\n"
	assert.Equal(t, resp2, body2)

}

func TestGqlJSONThrottle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}

	data, err := ioutil.ReadFile(filepath.FromSlash("./json-throttle-server-time/flogo.json"))
	assert.Nil(t, err)
	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)

	defer api.ClearResources()
	test.Drain("9096")
	err = e.Start()
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

	payload, _ := json.Marshal(&validPayload)

	request := func() string {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
		assert.Nil(t, err)
		response, err := client.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		response.Body.Close()
		return string(body)
	}

	body := request()
	resp := "{\"response\":\"{\\\"data\\\":{\\\"stationWithEvaId\\\":{\\\"name\\\":\\\"Frankfurt (Main) Hbf\\\"}}}\",\"validationMessage\":null}\n"
	assert.Equal(t, resp, body)
}
