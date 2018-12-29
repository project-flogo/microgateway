package examples

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"
	"strconv"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/api"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

// Response is a reply form the server
type Response struct {
	Status string `json:"status"`
}

func testApplication(t *testing.T, e engine.Engine, limit string) {
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

	num,_ := strconv.Atoi(string(limit[0]))
	for i := 0; i < num; i++ {
		response := request("TOKEN1")
		assert.NotEqual(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)
		assert.NotEqual(t, "Token not found", response.Status)
	}
	response := request("TOKEN1")
	assert.Equal(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)

	response = request("TOKEN2")
	assert.NotEqual(t, "Rate Limit Exceeded - The service you have requested is over the allowed limit.", response.Status)
	assert.NotEqual(t, "Token not found", response.Status)
	if string(limit[2]) == "M"{
		time.Sleep(time.Minute + time.Duration(num)*time.Second)
	}else if string(limit[2]) == "S"{
		time.Sleep(time.Second + time.Duration(num)*time.Second)
	}else{
		time.Sleep(time.Hour + time.Duration(num)*time.Second)
	}
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
	parameters := []struct {
		limit string
	}{
		{"3-M"},{"1-S"},
	}
	for i := range parameters {
		e, err := Example(parameters[i].limit)
		assert.Nil(t, err)
		testApplication(t, e,parameters[i].limit)
	}
}

func TestIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}
	parameters := []struct {
		limit string
	}{
		{"4-M"},{"2-S"},
	}
	data, err := ioutil.ReadFile(filepath.FromSlash("./json/flogo.json"))
	assert.Nil(t, err)
	for i := range parameters {
		var Input input
		err = json.Unmarshal(data, &Input)
		Input.Resources[0].Data.Services[0]["settings"] = map[string]interface{}{"limit": parameters[i].limit}
		data, _ = json.Marshal(Input)

		cfg, err := engine.LoadAppConfig(string(data), false)
		assert.Nil(t, err)
		e, err := engine.New(cfg)
		assert.Nil(t, err)
		testApplication(t, e, parameters[i].limit)
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