package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	trigger "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway"

	_ "github.com/project-flogo/contrib/activity/rest"
	_ "github.com/project-flogo/microgateway/activity/circuitbreaker"
	_ "github.com/project-flogo/microgateway/activity/jwt"
	_ "github.com/project-flogo/microgateway/activity/ratelimiter"
)

var (
	server = flag.Bool("server", false, "run the test server")
)

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

func main() {
	flag.Parse()

	if *server {
		http.HandleFunc("/pets", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("url: %q\n", html.EscapeString(r.URL.Path))
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(body))
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write([]byte(reply))
			if err != nil {
				panic(err)
			}
		})

		err := http.ListenAndServe(":1234", nil)
		if err != nil {
			panic(err)
		}

		return
	}

	app := api.NewApp()

	trg := app.NewTrigger(&trigger.Trigger{}, &trigger.Settings{Port: 9096})
	handler, err := trg.NewHandler(&trigger.HandlerSettings{
		Method: "GET",
		Path:   "/endpoint",
	})
	if err != nil {
		panic(err)
	}

	_, err = handler.NewAction(&microgateway.Action{}, map[string]interface{}{
		"pattern":           "DefaultHttpPattern",
		"useRateLimiter":    true,
		"rateLimit":         "1-S",
		"useJWT":            true,
		"jwtSigningMethod":  "HMAC",
		"jwtKey":            "qwertyuiopasdfghjklzxcvbnm789101",
		"jwtAud":            "www.mashling.io",
		"jwtIss":            "Mashling",
		"jwtSub":            "tempuser@mail.com",
		"useCircuitBreaker": true,
		"backendUrl":        "http://localhost:1234/pets",
	})
	if err != nil {
		panic(err)
	}

	e, err := api.NewEngine(app)
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
