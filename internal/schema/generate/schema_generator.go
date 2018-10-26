// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mashling-support/jsonschema"
	"github.com/project-flogo/microgateway/api"
)

func main() {
	schema := jsonschema.Reflect(&api.Microgateway{})
	schemaJSON, err := json.MarshalIndent(schema, "", "    ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile("schema.json", schemaJSON, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
