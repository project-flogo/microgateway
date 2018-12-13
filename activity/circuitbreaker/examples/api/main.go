package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/circuitbreaker/examples"
)

func main() {
	e, err := examples.Example("a",2,0,0)
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
