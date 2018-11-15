package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
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

var (
	server = flag.Bool("server", false, "run the test server")
	client = flag.Bool("client", false, "run the test client")
)

// Response is a reply form the server
type Response struct {
	Error      string  `json:"error"`
	Complexity float64 `json:"complexity"`
}

func main() {
	flag.Parse()

	if *client {
		count, sum := 0, 0.0
		rnd := rand.New(rand.NewSource(1))
		client := &http.Client{}
		for i := 0; i < 1024; i++ {
			data, err := json.Marshal(generateRandomJSON(rnd))
			if err != nil {
				panic(err)
			}
			req, err := http.NewRequest(http.MethodPut, "http://localhost:9096/test", bytes.NewReader(data))
			if err != nil {
				panic(err)
			}
			req.Header.Set("Content-Type", "application/json")
			response, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				panic(err)
			}
			response.Body.Close()
			rsp := Response{}
			fmt.Println("body", string(body))
			err = json.Unmarshal(body, &rsp)
			if err != nil {
				panic(err)
			}
			if rsp.Error == "anomaly!" {
				count++
			}
			sum += rsp.Complexity
		}
		fmt.Println("number of anomalies", count)
		fmt.Println("average complexity", sum/float64(count))
	}

	if *server {
		http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("url: %q\n", html.EscapeString(r.URL.Path))
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(body))
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(body)
			if err != nil {
				panic(err)
			}
		})

		err := http.ListenAndServe(":1234", nil)
		if err != nil {
			panic(err)
		}
	}
}
