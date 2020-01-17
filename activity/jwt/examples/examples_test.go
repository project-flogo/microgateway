package examples

import (
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
	Pet   json.RawMessage `json:"pet"`
	Error json.RawMessage `json:"error"`
}

type Error struct {
	ValidationMessage string `json:"validationMessage"`
}

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

	request := func(auth string) Response {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets", nil)
		assert.Nil(t, err)
		if auth != "" {
			req.Header.Add("Authorization", auth)
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

	response := request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU3OTI1MDE2NiwiZXhwIjoxNjEwNzg2MTY2LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsIkdpdmVuTmFtZSI6IkpvaG5ueSIsIlN1cm5hbWUiOiJSb2NrZXQiLCJFbWFpbCI6Impyb2NrZXRAZXhhbXBsZS5jb20iLCJSb2xlIjpbIk1hbmFnZXIiLCJQcm9qZWN0IEFkbWluaXN0cmF0b3IiXX0.IUlvLDRYPifc3lR2X331NQUxiZaEldtr5DUzLI7Zsj4")
	assert.Equal(t, "\"JWT token is valid\"", string(response.Error))
	assert.Condition(t, func() bool {
		return len(response.Pet) > 0
	})

	response = request("Bearer <Access_Token>")
	var er Error
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "token contains an invalid number of segments", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))

	response = request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0MDQ4NzgzNywiZXhwIjoxNTcyMDIzODM4LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsImlkIjoiMSJ9.-Tzfn5ZS0kM-u07qkpFrDxdyptBJIvLesuUzVXdqn4")
	er = Error{}
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "signature is invalid", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))

	response = request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZ0Vycm9yIiwiaWF0IjoxNTQ0MDUxNTIwLCJleHAiOjE1NzU1ODc1MjEsImF1ZCI6Ind3dy5tYXNobGluZy5pbyIsInN1YiI6InRlbXB1c2VyQG1haWwuY29tIiwiaWQiOiI0Iiwic2lnbmluZ01ldGhvZCI6IkhNQUMifQ.r4yFp-UEBf7gniNI4A2dAUa8kQgPlowI5hrgnwsFdd8")
	er = Error{}
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "iss claims do not match", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))

	response = request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDA1MTUyMCwiZXhwIjoxNTc1NTg3NTIxLCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8uZXJyb3IiLCJzdWIiOiJ0ZW1wdXNlckBtYWlsLmNvbSIsImlkIjoiNCIsInNpZ25pbmdNZXRob2QiOiJITUFDIn0.Fp2-O5fuO5b9r3DBafN70AbkenLn3gJuikjRZNxaY0M")
	er = Error{}
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "aud claims do not match", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))

	response = request("Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTU0NDA1MTUyMCwiZXhwIjoxNTc1NTg3NTIxLCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJ0ZW1wdXNlcmVycm9yQG1haWwuY29tIiwiaWQiOiI0Iiwic2lnbmluZ01ldGhvZCI6IkhNQUMifQ.S99Pr15zK3ZLAEkZZ9ObcL6VAczrX6ojZRUvOcH3RPo")
	er = Error{}
	err = json.Unmarshal(response.Error, &er)
	assert.Nil(t, err)
	assert.Equal(t, "sub claims do not match", er.ValidationMessage)
	assert.Equal(t, "null", string(response.Pet))
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
