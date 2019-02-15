package main

import (

	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/microgateway/examples"
	_ "github.com/project-flogo/microgateway/activity/circuitbreaker"
	_ "github.com/project-flogo/microgateway/activity/jwt"
	_ "github.com/project-flogo/contrib/activity/channel"
)

func main() {

	e, err := examples.DefaultChannelPattern()
	if err != nil {
		panic(err)
	}
	engine.RunEngine(e)
}
