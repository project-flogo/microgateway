package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var client = flag.Bool("client", false, "run the client")

func main() {
	flag.Parse()

	if *client {
		client := &http.Client{}
		request := func() []byte {
			req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/pets/1", nil)
			if err != nil {
				panic(err)
			}
			req.Header.Add("Token", "ABC123")
			response, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				panic(err)
			}
			response.Body.Close()

			return body
		}
		count := 0
		for i := 0; i < 256; i++ {
			time.Sleep(50 * time.Millisecond)
			output := request()
			fmt.Println(count, string(output))
			count++
		}
		wait := sync.WaitGroup{}
		wait.Add(10)
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			go func(count int) {
				output := request()
				fmt.Println(count, string(output))
				wait.Done()
			}(count)
			count++
		}
		wait.Wait()
		for i := 0; i < 256; i++ {
			time.Sleep(50 * time.Millisecond)
			output := request()
			fmt.Println(count, string(output))
			count++
		}
		return
	}
}
