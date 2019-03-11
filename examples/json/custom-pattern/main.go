package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	"github.com/project-flogo/microgateway"
)

func init() {
	data, err := ioutil.ReadFile("/Users/agadikar/microgateway/examples/json/custom-pattern/CustomPattern.json")
	if err != nil {
		panic(err)
	}
	err = microgateway.Register("CustomPattern", string(data))
	if err != nil {
		panic(err)
	}
}

var (
	server = flag.Bool("server", false, "run the test server")
)

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
			_, err = w.Write([]byte(""))
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
}
