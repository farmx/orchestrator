package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetConstant(t *testing.T) {
	headerKey := "HEADER_KEY"
	headerValue := "HEADER_VALUE"
	headerKey2 := "HEADER_KEY_2"
	headerValue2 := "HEADER_VALUE_2"

	ctx, _ := NewContext()
	if err := ctx.setConstant(headerKey, headerValue); err != nil {
		t.Fail()
	}

	if err := ctx.setConstant(headerKey2, headerValue2); err != nil {
		t.Fail()
	}

	if err := ctx.setConstant(headerKey2, headerValue); err == nil {
		t.Fail()
	}

	assert.NotNil(t, ctx.getGid())
	assert.Equal(t, ctx.getConstant(headerKey), headerValue)
	assert.Equal(t, ctx.getConstant(headerKey2), headerValue2)
}

func TestSetVariable(t *testing.T) {
	headerKey := "HEADER_KEY"
	headerValue := "HEADER_VALUE"
	headerKey2 := "HEADER_KEY_2"
	headerValue2 := "HEADER_VALUE_2"

	ctx, _ := NewContext()
	ctx.setVariable(headerKey, headerValue)
	ctx.setVariable(headerKey2, headerValue2)
	ctx.setVariable(headerKey2, headerValue)

	assert.NotNil(t, ctx.getGid())
	assert.Equal(t, headerValue, ctx.getVariable(headerKey))
	assert.Equal(t, headerValue, ctx.getVariable(headerKey2))
}
