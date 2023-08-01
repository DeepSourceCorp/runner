package httperror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(500, "message", errors.New("error"))
	err2 := errors.Unwrap(err)
	assert.EqualError(t, err2, "error")
}
