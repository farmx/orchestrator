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
	if err := ctx.SetConstant(headerKey, headerValue); err != nil {
		t.Fail()
	}

	if err := ctx.SetConstant(headerKey2, headerValue2); err != nil {
		t.Fail()
	}

	if err := ctx.SetConstant(headerKey2, headerValue); err == nil {
		t.Fail()
	}

	assert.NotNil(t, ctx.GetGid())
	assert.Equal(t, ctx.GetConstant(headerKey), headerValue)
	assert.Equal(t, ctx.GetConstant(headerKey2), headerValue2)
}

func TestSetVariable(t *testing.T) {
	headerKey := "HEADER_KEY"
	headerValue := "HEADER_VALUE"
	headerKey2 := "HEADER_KEY_2"
	headerValue2 := "HEADER_VALUE_2"

	ctx, _ := NewContext()
	ctx.SetVariable(headerKey, headerValue)
	ctx.SetVariable(headerKey2, headerValue2)
	ctx.SetVariable(headerKey2, headerValue)

	assert.NotNil(t, ctx.GetGid())
	assert.Equal(t, headerValue, ctx.GetVariable(headerKey))
	assert.Equal(t, headerValue, ctx.GetVariable(headerKey2))
}
