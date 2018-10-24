package error

import (
	"reflect"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
)

func init() {
	function.Register(&fnType{})
}

type fnType struct {
}

// Name returns the name of the function
func (fnType) Name() string {
	return "error.type"
}

// Sig returns the function signature
func (fnType) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeAny}, false
}

// Eval executes the function
func (fnType) Eval(params ...interface{}) (interface{}, error) {
	return reflect.TypeOf(params[0]).String(), nil
}
