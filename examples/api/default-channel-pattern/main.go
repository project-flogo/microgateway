package main

import (
	_ "github.com/project-flogo/contrib/activity/channel"
	"github.com/project-flogo/core/engine"
	_ "github.com/project-flogo/microgateway/activity/circuitbreaker"
	_ "github.com/project-flogo/microgateway/activity/jwt"
	"github.com/project-flogo/microgateway/examples"
)

func main() {

	e, err := examples.DefaultChannelPattern()
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
