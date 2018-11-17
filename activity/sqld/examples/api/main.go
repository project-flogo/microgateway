package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/sqld"
	"github.com/project-flogo/microgateway/activity/sqld/example"
)

func main() {
	e, err := example.Example(&sqld.Activity{})
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
