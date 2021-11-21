package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// S1--->S2
func TestHappyScenario(t *testing.T) {
	s1 := &State{
		name: "S1",
		action: func(ctx *context) error {
			log.Print("State 1 action")
			return nil
		},
	}

	s2 := &State{
		name: "S2",
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return nil
		},
	}

	s1.createTransition(s2, 1,
		func(ctx context) bool {
			return true
		})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(s1, ctx)

	// init time
	hasNext, _ := sm.doAction()
	assert.Equal(t, true, hasNext)
	assert.Equal(t, s2, sm.state)

	// move to getLabel state
	hasNext, _ = sm.doAction()
	assert.Equal(t, s2, sm.state)
	assert.Equal(t, false, hasNext)
}

// --->S1---->S2
// \__/
func TestLoop(t *testing.T) {
	const headerKey = "HEADER_KEY"

	s1 := &State{
		name: "S1",
		action: func(ctx *context) error {
			if v := ctx.GetVariable(headerKey); v == nil {
				ctx.SetVariable(headerKey, 0)
			}

			ctr := ctx.GetVariable(headerKey).(int) + 1
			log.Print("State 1 action")
			return ctx.SetVariable(headerKey, ctr)
		},
	}

	s2 := &State{
		name: "S2",
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return nil
		},
	}

	s1.createTransition(s2, 1,
		func(ctx context) bool {
			return ctx.GetVariable(headerKey) == 3
		})

	s1.createTransition(s1, 1,
		func(ctx context) bool {
			return ctx.GetVariable(headerKey).(int) < 3
		})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(s1, ctx)

	// cycle one
	hasNext, _ := sm.doAction()
	assert.Equal(t, true, hasNext)
	assert.Equal(t, s1, sm.state)

	// cycle two
	hasNext, _ = sm.doAction()
	assert.Equal(t, true, hasNext)
	assert.Equal(t, s1, sm.state)

	// cycle three
	hasNext, _ = sm.doAction()
	assert.Equal(t, true, hasNext)
	assert.Equal(t, s2, sm.state)

	// cycle four
	hasNext, _ = sm.doAction()
	assert.Equal(t, false, hasNext)
	assert.Equal(t, s2, sm.state)
}
//  /---->S3---------\
// S1--->S2--->S4--->S5
//  \----------------/
func TestComplexCondition(t *testing.T) {
	s1 := &State{
		name: "S1",
		action: func(ctx *context) error {
			log.Print("State 1 action")
			return nil
		},
	}

	s2 := &State{
		name: "S2",
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return nil
		},
	}

	s3 := &State{
		name: "S3",
		action: func(ctx *context) error {
			log.Print("State 3 action")
			return nil
		},
	}

	s4 := &State{
		name: "S4",
		action: func(ctx *context) error {
			log.Print("State 4 action")
			return nil
		},
	}

	s5 := &State{
		name: "S5",
		action: func(ctx *context) error {
			log.Print("State 5 action")
			return nil
		},
	}

	s1.createTransition(s2, 2,
		func(ctx context) bool {
			return true
		})

	s1.createTransition(s3, 2,
		func(ctx context) bool {
			return false
		})

	s2.createTransition(s4,1,
		func(ctx context) bool {
			return true
		})

	s4.createTransition(s5,1,
		func(ctx context) bool {
			return true
		})

	s3.createTransition(s5, 1,
		func(ctx context) bool {
			return true
		})

	s1.createTransition(s5, 1,
		func(ctx context) bool {
			return true
		})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(s1, ctx)

	for hasNext := false; hasNext != false; hasNext, _ = sm.doAction() {

	}
}
