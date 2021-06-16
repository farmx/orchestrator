package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContext_SetVariableWithVersion(t *testing.T) {
	key := "KEY"
	v1 := "v1"
	v2 := "v2"
	value1 := "VALUE_1"
	value2 := "VALUE_2"

	ctx, _ := NewContext()
	if err := ctx.SetVariableWithVersion(key, v1, v1, value1); err != nil {
		t.Fail()
	}

	if err := ctx.SetVariableWithVersion(key, v1, v2, value2); err != nil {
		t.Fail()
	}

	if err := ctx.SetVariableWithVersion(key, v1, v1, value2); err == nil {
		t.Fail()
	}

	assert.NotNil(t, ctx.GetGid())
	assert.Equal(t, ctx.GetVariable(key), value2)
}

func TestContext_SetVariable(t *testing.T) {
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
