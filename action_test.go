package microgateway

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/project-flogo/contrib/activity/rest"
	coreactivity "github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/api"
	microapi "github.com/project-flogo/microgateway/api"
	"github.com/project-flogo/microgateway/internal/testing/activity"
	"github.com/project-flogo/microgateway/internal/testing/trigger"
	"github.com/stretchr/testify/assert"

	_ "github.com/project-flogo/contrib/activity/channel"
	_ "github.com/project-flogo/contrib/activity/rest"
	_ "github.com/project-flogo/microgateway/activity/circuitbreaker"
	_ "github.com/project-flogo/microgateway/activity/jwt"
	_ "github.com/project-flogo/microgateway/activity/ratelimiter"
)

func TestMicrogateway(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()
	app := api.NewApp()

	microgateway := microapi.New("test")
	service := microgateway.NewService("test", &activity.Activity{})
	service.SetDescription("A test activity")
	service.AddSetting("message", "hello world")
	step := microgateway.NewStep(service)
	step.SetIf("1 == 1")
	step.AddInput("message", "=1337")
	response := microgateway.NewResponse(false)
	response.SetCode("=200")
	response.SetData(map[string]interface{}{
		"test": "=$.test.outputs.data",
		"foo":  "bar",
		"bar":  1.0,
	})
	settings, err := microgateway.AddResource(app)
	assert.Nil(t, err)

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, settings)
	assert.Nil(t, err)
	action.SetCondition(`$.content.a == "b"`)

	defaultActionHit := false
	action, err = handler.NewAction(func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
		defaultActionHit = true
		return nil, nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()
	defer e.Stop()

	result, err := trigger.Fire(0, map[string]interface{}{"a": "c"})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, "", activity.Message)
	assert.False(t, activity.HasEvaled)
	assert.True(t, defaultActionHit)
	defaultActionHit = false

	result, err = trigger.Fire(0, map[string]interface{}{"a": "b"})
	assert.Nil(t, err)
	assert.Equal(t, 200, result["code"])
	assert.Equal(t, "1337", result["data"].(map[string]interface{})["test"])
	assert.Equal(t, "bar", result["data"].(map[string]interface{})["foo"])
	assert.Equal(t, 1.0, result["data"].(map[string]interface{})["bar"])
	assert.Equal(t, "1337", activity.Message)
	assert.True(t, activity.HasEvaled)
	assert.False(t, defaultActionHit)
}

func TestMicrogatewayHalt(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()
	app := api.NewApp()

	microgateway := microapi.New("halt")
	serviceHalt := microgateway.NewService("halt", &rest.Activity{})
	serviceHalt.SetDescription("An activity that will halt")
	serviceHalt.AddSetting("uri", "http://localhost:1234/abc123")
	serviceHalt.AddSetting("method", "GET")
	serviceTest := microgateway.NewService("test", &activity.Activity{})
	serviceTest.SetDescription("A test activity")
	serviceTest.AddSetting("message", "hello world")
	step := microgateway.NewStep(serviceHalt)
	step.SetHalt("($.halt.error != nil) && !error.isneterror($.halt.error)")
	step = microgateway.NewStep(serviceTest)
	assert.NotNil(t, step)
	response := microgateway.NewResponse(true)
	response.SetCode("=403")
	response.SetData(map[string]interface{}{
		"isneterror": "=error.isneterror($.halt.error)",
		"error":      "=error.string($.halt.error)",
	})
	response = microgateway.NewResponse(false)
	response.SetCode("=200")
	response.SetData(map[string]interface{}{
		"message": "hello world",
	})
	settings, err := microgateway.AddResource(app)
	assert.Nil(t, err)

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, settings)
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()
	defer e.Stop()

	result, err := trigger.Fire(0, map[string]interface{}{})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, true, result["data"].(map[string]interface{})["isneterror"])
	assert.True(t, activity.HasEvaled)
}

func TestMicrogatewayHandler(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()
	app := api.NewApp()

	microgateway := microapi.New("test")
	fired, message := false, ""
	service := microgateway.NewService("test", func(ctx coreactivity.Context) (done bool, err error) {
		fired = true
		message = fmt.Sprintf("%v", ctx.GetInput("message"))
		ctx.SetOutput("data", message)
		return true, nil
	})
	service.SetDescription("A test activity")
	service.AddSetting("message", "hello world")
	step := microgateway.NewStep(service)
	step.SetIf("1 == 1")
	step.AddInput("message", "=1337")
	response := microgateway.NewResponse(false)
	response.SetCode("=200")
	response.SetData(map[string]interface{}{
		"test": "=$.test.outputs.data",
		"foo":  "bar",
		"bar":  1.0,
	})
	settings, err := microgateway.AddResource(app)
	assert.Nil(t, err)

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, settings)
	assert.Nil(t, err)
	action.SetCondition(`$.content.a == "b"`)

	defaultActionHit := false
	action, err = handler.NewAction(func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
		defaultActionHit = true
		return nil, nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()
	defer e.Stop()

	result, err := trigger.Fire(0, map[string]interface{}{"a": "c"})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, "", message)
	assert.False(t, fired)
	assert.True(t, defaultActionHit)
	defaultActionHit = false

	result, err = trigger.Fire(0, map[string]interface{}{"a": "b"})
	assert.Nil(t, err)
	assert.Equal(t, 200, result["code"])
	assert.Equal(t, "1337", result["data"].(map[string]interface{})["test"])
	assert.Equal(t, "bar", result["data"].(map[string]interface{})["foo"])
	assert.Equal(t, 1.0, result["data"].(map[string]interface{})["bar"])
	assert.Equal(t, "1337", message)
	assert.True(t, fired)
	assert.False(t, defaultActionHit)
}

type handler struct {
	hit bool
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hit = true
	r.Body.Close()
}

func TestMicrogatewayHttpPattern(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()

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
	_, err := http.Get("http://localhost:1234/")
	for err != nil {
		_, err = http.Get("http://localhost:1234/")
	}
	defer s.Shutdown(context.Background())
	testHandler.hit = false

	app := api.NewApp()

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, map[string]interface{}{
		"pattern":           "DefaultHttpPattern",
		"useRateLimiter":    false,
		"useJWT":            false,
		"useCircuitBreaker": false,
		"backendUrl":        "http://localhost:1234/",
		"rateLimit":         "3-M",
	})
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()
	defer e.Stop()

	result, err := trigger.Fire(0, map[string]interface{}{})
	assert.Nil(t, err)
	assert.NotNil(t, result)

	assert.True(t, testHandler.hit)
}

func TestMicrogatewayChannelPattern(t *testing.T) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()

	app := api.NewApp()

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	assert.Nil(t, err)

	action, err := handler.NewAction(&Action{}, map[string]interface{}{
		"pattern":           "DefaultChannelPattern",
		"useJWT":            false,
		"useCircuitBreaker": false,
	})
	assert.Nil(t, err)
	assert.NotNil(t, action)

	e, err := api.NewEngine(app)
	assert.Nil(t, err)
	e.Start()
	defer e.Stop()

	result, err := trigger.Fire(0, map[string]interface{}{})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func BenchmarkMicrogateway(b *testing.B) {
	defer func() {
		trigger.Reset()
		activity.Reset()
	}()
	app := api.NewApp()

	microgateway := microapi.New("benchmark")
	service := microgateway.NewService("test", &activity.Activity{})
	service.SetDescription("A benchmark activity")
	service.AddSetting("message", "hello world")
	for i := 0; i < 256; i++ {
		step := microgateway.NewStep(service)
		if step == nil {
			b.Fatal("failed to create step")
		}
	}
	response := microgateway.NewResponse(false)
	response.SetCode("=200")
	response.SetData(map[string]interface{}{
		"foo": "bar",
	})
	settings, err := microgateway.AddResource(app)
	if err != nil {
		b.Fatal(err)
	}

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{ASetting: 1337})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{})
	if err != nil {
		b.Fatal(err)
	}

	action, err := handler.NewAction(&Action{}, settings)
	if err != nil {
		b.Fatal(err)
	}
	if action == nil {
		b.Fatal("failed to create action")
	}

	e, err := api.NewEngine(app)
	if err != nil {
		b.Fatal(err)
	}
	if e == nil {
		b.Fatal("failed to create app engine")
	}
	e.Start()
	defer e.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trigger.Fire(0, map[string]interface{}{})
	}
}
