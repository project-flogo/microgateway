package examples

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/api"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

// PetStore is a petstore
type PetStore struct {
	t *testing.T
}

const petStoreResponse = `{
	"id": 1,
	"category": {
		"id": 0,
		"name": "string"
	},
	"name": "sally",
	"photoUrls": ["string"],
	"tags": [{"id": 0,"name": "string"}],
	"status": "available"
}
`

// ServeHTTP handle a petstore request
func (p *PetStore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(petStoreResponse))
	if err != nil {
		p.t.Fatal(err)
	}
}

// StartPetStore starts the petstore
func StartPetStore(t *testing.T) *http.Server {
	server := http.Server{
		Addr: ":8080",
		Handler: &PetStore{
			t: t,
		},
	}
	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}()
	return &server
}

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

	test.Drain("8080")
	store := StartPetStore(t)
	defer store.Shutdown(context.Background())
	test.Pour("8080")

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

	num, _ := strconv.Atoi(string(limit[0]))
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
	if string(limit[2]) == "M" {
		time.Sleep(time.Minute + time.Duration(num)*time.Second)
	} else if string(limit[2]) == "S" {
		time.Sleep(time.Second + time.Duration(num)*time.Second)
	} else {
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
		{"3-M"}, {"1-S"},
	}
	for i := range parameters {
		e, err := Example(parameters[i].limit, 0)
		assert.Nil(t, err)
		testApplication(t, e, parameters[i].limit)
	}
}

func TestIntegrationJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}
	parameters := []struct {
		limit string
	}{
		{"3-M"}, {"1-S"},
	}
	data, err := ioutil.ReadFile(filepath.FromSlash("./json/basic/flogo.json"))
	assert.Nil(t, err)
	for i := range parameters {
		var Input input
		err = json.Unmarshal(data, &Input)
		assert.Nil(t, err)
		Input.Resources[0].Data.Services[0]["settings"] = map[string]interface{}{"limit": parameters[i].limit}
		Input.Resources[0].Data.Services[1]["settings"].(map[string]interface{})["uri"] = "http://localhost:8080/v2/pet/:petId"
		data, _ = json.Marshal(Input)

		cfg, err := engine.LoadAppConfig(string(data), false)
		assert.Nil(t, err)
		e, err := engine.New(cfg)
		assert.Nil(t, err)
		testApplication(t, e, parameters[i].limit)
	}
}

func testSmartApplication(t *testing.T, e engine.Engine) {
	const failCondition = "Rate Limit Exceeded - The service you have requested is over the allowed limit."
	defer api.ClearResources()
	test.Drain("9096")
	err := e.Start()
	assert.Nil(t, err)
	defer func() {
		err := e.Stop()
		assert.Nil(t, err)
	}()
	test.Pour("9096")

	test.Drain("8080")
	store := StartPetStore(t)
	defer store.Shutdown(context.Background())
	test.Pour("8080")

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

	for i := 0; i < 256; i++ {
		time.Sleep(50 * time.Millisecond)
		response := request("TEST")
		assert.NotEqual(t, response.Status, failCondition)
	}
	wait, blocked, notBlocked := sync.WaitGroup{}, uint64(0), uint64(0)
	wait.Add(10)
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Millisecond)
		go func() {
			response := request("TEST")
			if response.Status == failCondition {
				atomic.AddUint64(&blocked, 1)
			} else {
				atomic.AddUint64(&notBlocked, 1)
			}
			wait.Done()
		}()
	}
	wait.Wait()
	for i := 0; i < 256; i++ {
		time.Sleep(50 * time.Millisecond)
		response := request("TEST")
		if response.Status == failCondition {
			blocked++
		} else {
			notBlocked++
		}
	}
	assert.Condition(t, func() (success bool) {
		return blocked > 0
	}, "some requests should have been blocked")
	assert.Condition(t, func() (success bool) {
		return notBlocked > 0
	}, "some requests should not have been blocked")
}

func TestIntegrationSmartAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping API integration test in short mode")
	}
	e, err := Example("1000-S", 2)
	assert.Nil(t, err)
	testSmartApplication(t, e)
}

func TestIntegrationSmartJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JSON integration test in short mode")
	}
	data, err := ioutil.ReadFile(filepath.FromSlash("./json/smart/flogo.json"))
	assert.Nil(t, err)

	var Input input
	err = json.Unmarshal(data, &Input)
	assert.Nil(t, err)
	Input.Resources[0].Data.Services[1]["settings"].(map[string]interface{})["uri"] = "http://localhost:8080/v2/pet/:petId"
	data, _ = json.Marshal(Input)

	cfg, err := engine.LoadAppConfig(string(data), false)
	assert.Nil(t, err)
	e, err := engine.New(cfg)
	assert.Nil(t, err)
	testSmartApplication(t, e)
}

//--------data structure-------//

type input struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Version   string      `json:"version"`
	Desc      string      `json:"description"`
	Prop      interface{} `json:"properties"`
	Channels  interface{} `json:"channels"`
	Trig      interface{} `json:"triggers"`
	Resources []struct {
		ID       string `json:"id"`
		Compress bool   `json:"compressed"`
		Data     struct {
			Name      string                   `json:"name"`
			Steps     []interface{}            `json:"steps"`
			Responses []interface{}            `json:"responses"`
			Services  []map[string]interface{} `json:"services"`
		} `json:"data"`
	} `json:"resources"`
	Actions interface{} `json:"actions"`
}
