package orchestrator

import (
	"strconv"
)

type labelGenerator struct {
	prefix   string
	versions []int
}

func newLabelGenerator(prefix string) *labelGenerator {
	return &labelGenerator{
		prefix:   prefix,
		versions: []int{0},
	}
}

func (c *labelGenerator) getLabel() string {
	index := len(c.versions) - 1
	c.versions[index] = c.versions[index] + 1

	cr := strconv.Itoa(c.versions[0])
	for _, v := range c.versions[1:] {
		cr = cr + "." + strconv.Itoa(v)
	}

	return cr
	//return fmt.Sprintf("%s_%s", c.prefix, cr)
}

func (c *labelGenerator) hasChild() {
	c.versions = append(c.versions, 0)
}

func (c *labelGenerator) endChild() {
	lenv := len(c.versions)
	c.versions = c.versions[:lenv-1]
}
