package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounting(t *testing.T) {
	c := newCounter()

	r1 := c.next()
	assert.Equal(t, "1", r1)

	r2 := c.next()
	assert.Equal(t, "2", r2)

	c.subVersioning()
	r3 := c.next()
	assert.Equal(t, "2.1", r3)

	r4 := c.next()
	assert.Equal(t, "2.2", r4)

	c.endSubVersioning()
	r5 := c.next()
	assert.Equal(t, "3", r5)
}
