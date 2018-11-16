package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/circuitbreaker"
	"github.com/project-flogo/microgateway/activity/circuitbreaker/example"
)

func main() {
	e, err := example.Example(&circuitbreaker.Activity{})
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
