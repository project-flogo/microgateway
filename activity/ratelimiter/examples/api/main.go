package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/ratelimiter"
	"github.com/project-flogo/microgateway/activity/ratelimiter/example"
)

func main() {
	e, err := example.Example(&ratelimiter.Activity{})
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
