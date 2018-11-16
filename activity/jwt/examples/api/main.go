package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/jwt"
	"github.com/project-flogo/microgateway/activity/jwt/example"
)

func main() {
	e, err := example.Example(&jwt.Activity{})
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
