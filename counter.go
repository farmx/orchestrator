package orchestrator

import (
	"strconv"
)

type counter struct {
	versions []int
}

func newCounter() *counter {
	return &counter{
		versions: []int{0},
	}
}

func (c *counter) next() string {
	index := len(c.versions) - 1
	c.versions[index] = c.versions[index] + 1

	cr := strconv.Itoa(c.versions[0])
	for _, v := range c.versions[1:] {
		cr = cr + "." + strconv.Itoa(v)
	}

	return cr
}

func (c *counter) subVersioning() {
	c.versions = append(c.versions, 0)
}

func (c *counter) endSubVersioning() {
	lenv := len(c.versions)
	c.versions = c.versions[:lenv-1]
}
