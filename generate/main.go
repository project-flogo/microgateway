package main

import (
	"io/ioutil"

	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway"
)

func main() {
	flogo, err := ioutil.ReadFile("./flogo.json")
	if err != nil {
		panic(err)
	}
	app, err := engine.LoadAppConfig(string(flogo), false)
	if err != nil {
		panic(err)
	}
	api.Generate(app, "app.go")
}
