package error

import (
	"errors"
	"testing"

	"github.com/project-flogo/core/data/expression/function"
	"github.com/stretchr/testify/assert"
)

func TestFnString_Eval(t *testing.T) {
	f := &fnString{}
	err1 := errors.New("test error")
	v, err := function.Eval(f, err1)
	assert.Nil(t, err)
	assert.Equal(t, "test error", v)
}
