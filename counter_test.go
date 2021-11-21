package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounting(t *testing.T) {
	c := newLabelGenerator("")

	r1 := c.getLabel()
	assert.Equal(t, "1", r1)

	r2 := c.getLabel()
	assert.Equal(t, "2", r2)

	c.hasChild()
	r3 := c.getLabel()
	assert.Equal(t, "2.1", r3)

	r4 := c.getLabel()
	assert.Equal(t, "2.2", r4)

	c.endChild()
	r5 := c.getLabel()
	assert.Equal(t, "3", r5)
}
