package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewContext(t *testing.T) {
	body := "BODY_CONTENT"
	headerKey := "HEADER_KEY"
	headerValue := "HEADER_VALUE"
	headerKey2 := "HEADER_KEY_2"
	headerValue2 := "HEADER_VALUE_2"

	ctx, _ := NewContext(body)
	ctx.setHeader(headerKey, headerValue)
	ctx.setHeader(headerKey2, headerValue2)

	assert.Equal(t, body, ctx.getBody())
	assert.NotNil(t, ctx.getGid())
	assert.Equal(t, ctx.getHeader(headerKey), headerValue)
	assert.Equal(t, ctx.getHeader(headerKey2), headerValue2)
}
