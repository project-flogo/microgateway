package main

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/activity/anomaly"
)

func main() {
	e, err := anomaly.Example()
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
