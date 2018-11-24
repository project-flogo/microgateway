package examples

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/project-flogo/core/engine"
	test "github.com/project-flogo/microgateway/internal/testing"
	"github.com/stretchr/testify/assert"
)

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

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
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

func testApplication(t *testing.T, e engine.Engine) {
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
	rnd, client := rand.New(rand.NewSource(1)), &http.Client{
		Transport: transport,
	}
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
		assert.Condition(t, func() bool {
			return rsp.Complexity > 8
		})
	}
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
