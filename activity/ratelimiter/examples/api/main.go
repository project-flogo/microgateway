package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/ratelimiter/examples"
)

func main() {
	e, err := examples.Example("3-M")
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
