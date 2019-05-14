package main

import (
	"flag"
	"io/ioutil"

	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway"
)

var (
	input  = flag.String("input", "flogo.json", "the input file")
	output = flag.String("output", "app.go", "the output file")
)

func main() {
	flag.Parse()

	flogo, err := ioutil.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	app, err := engine.LoadAppConfig(string(flogo), false)
	if err != nil {
		panic(err)
	}
	api.Generate(app, *output)
}
