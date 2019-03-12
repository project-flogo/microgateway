package custom_pattern

import (
	"io/ioutil"
	"github.com/project-flogo/microgateway"
)

func init() {
	data, err := ioutil.ReadFile("/Users/agadikar/microgateway/examples/json/custom-pattern/CustomPattern.json")
	if err != nil {
		panic(err)
	}
	err = microgateway.Register("CustomPattern", string(data))
	if err != nil {
		panic(err)
	}
}
