package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/examples"
)

func main() {
	e, err := examples.AsyncGatewayExample()
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
