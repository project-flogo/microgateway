package error

import (
	"errors"
	"net"
	"testing"

	"github.com/project-flogo/core/data/expression/function"
	"github.com/stretchr/testify/assert"
)

type testError struct {
	error
}

func (*testError) Timeout() bool {
	return false
}

func (*testError) Temporary() bool {
	return false
}

func TestFnIsNetError_Eval(t *testing.T) {
	f := &fnIsNetError{}
	var err1 net.Error = &testError{}
	v, err := function.Eval(f, err1)
	assert.Nil(t, err)
	assert.Equal(t, true, v)

	v, err = function.Eval(f, nil)
	assert.Nil(t, err)
	assert.Equal(t, false, v)

	err2 := errors.New("test error")
	v, err = function.Eval(f, err2)
	assert.Nil(t, err)
	assert.Equal(t, false, v)
}
