package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	_ "github.com/project-flogo/contrib/activity/rest"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway/activity/circuitbreaker"
	_ "github.com/project-flogo/contrib/activity/counter"
	"github.com/project-flogo/microgateway/examples"
)

var (
	server = flag.Bool("server", false, "run the test server")
)

const reply = `{
	"name": "sally"
}`

const pattern = `{
  "name": "CustomPattern",
  "steps": [
    {
      "service": "HttpBackend",
      "halt": "($.HttpBackend.error != nil) && !error.isneterror($.HttpBackend.error)"
    },
    {
      "if": "$.HttpBackend.error == nil",
      "service": "SuccessCounter"
    },
    {
      "if": "$.HttpBackend.error != nil",
      "service": "ErrorCounter"
    },
    {
      "service": "GetCounterSuccess"
    },
    {
      "service": "GetCounterError"
    }
  ],
  "responses": [
    {
      "if" : "$.GetCounterSuccess.error == nil",
      "error": false,
      "output": {
        "code": 200,
        "data": {
        	"Success-Calls": "=$.GetCounterSuccess.outputs.value",
        	"Error-Calls": "=$.GetCounterError.outputs.value"
        }
      }
    },
    {
      "error": true,
      "output": {
        "code": 400,
        "data": "Error"
      }
    }
  ],
  "services": [
    {
      "name": "HttpBackend",
      "description": "Make an http call to your backend",
      "ref": "github.com/project-flogo/contrib/activity/rest",
      "settings": {
        "method": "GET",
        "uri": "http://localhost:1234/pets"
      }
    },
    {
      "name": "SuccessCounter",
      "description": "Increment counter on successful call",
      "ref": "github.com/project-flogo/contrib/activity/counter",
      "settings": {
        "counterName": "SuccessCounter",
        "op": "increment"
      }
    },
    {
      "name": "ErrorCounter",
      "description": "Increment counter on error call",
      "ref": "github.com/project-flogo/contrib/activity/counter",
      "settings": {
        "counterName": "ErrorCounter",
        "op": "increment"
      }
    },
    {
      "name": "GetCounterSuccess",
      "description": "Get success counter",
      "ref": "github.com/project-flogo/contrib/activity/counter",
      "settings": {
        "counterName": "SuccessCounter"
      }
    },
    {
      "name": "GetCounterError",
      "description": "Get error counter",
      "ref": "github.com/project-flogo/contrib/activity/counter",
      "settings": {
        "counterName": "ErrorCounter"
      }
    }
  ]
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

	e, err := examples.CustomPattern("CustomPattern", pattern)
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
