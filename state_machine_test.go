package orchestrator

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestHappyScenario(t *testing.T) {
	state1 := &State{
		action: func(ctx *context) error {
			log.Print("State 1 action")
			return nil
		},
	}

	state2 := &State{
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return nil
		},
	}

	state1.transitions = append(state1.transitions, Transition{
		to:       state2,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, ctx)

	// init time
	assert.Equal(t, true, sm.hastNext())
	assert.Equal(t, state1, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))

	// cycle one
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))
	assert.Equal(t, true, sm.hastNext())

	// cycle three
	assert.Nil(t, sm.next())
	assert.Equal(t, false, sm.hastNext())
	assert.Equal(t, state2, sm.state)
	assert.Equal(t, SMEnd, sm.context.GetVariable(SMStatusHeaderKey))
}

func TestRollback(t *testing.T) {
	state1 := &State{
		action: func(ctx *context) error {
			if ctx.GetVariable(SMStatusHeaderKey) == SMRollback {
				log.Print("rollback call")
				return nil
			}

			log.Print("State 1 action")
			return nil
		},
	}

	state2 := &State{
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return errors.New("fake error")
		},
	}

	state1.transitions = append(state1.transitions, Transition{
		to:       state2,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) != SMRollback
		},
	})

	state2.transitions = append(state2.transitions, Transition{
		to:       state1,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) == SMRollback
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, ctx)

	// init time
	assert.Equal(t, true, sm.hastNext())
	assert.Equal(t, state1, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))

	// cycle one
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))
	assert.Equal(t, true, sm.hastNext())

	// cycle two
	assert.NotNil(t, sm.next())
	assert.Equal(t, state1, sm.state)
	assert.Equal(t, SMRollback, sm.context.GetVariable(SMStatusHeaderKey))
	assert.Equal(t, true, sm.hastNext())

	// cycle three
	assert.Nil(t, sm.next())
	assert.Equal(t, false, sm.hastNext())
	assert.Equal(t, state1, sm.state)
	assert.Equal(t, SMEnd, sm.context.GetVariable(SMStatusHeaderKey))
}

func TestLoop(t *testing.T) {
	headerKey := "HEADER_KEY"

	state1 := &State{
		action: func(ctx *context) error {
			if v := ctx.GetVariable(headerKey); v == nil {
				ctx.SetVariable(headerKey, 0)
			}

			ctx.SetVariable(headerKey, ctx.GetVariable(headerKey).(int)+1)
			log.Print("State 1 action")
			return nil
		},
	}

	state2 := &State{
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return nil
		},
	}

	state1.transitions = append(state1.transitions, Transition{
		to:       state2,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.GetVariable(headerKey) == 3
		},
	})

	state1.transitions = append(state1.transitions, Transition{
		to:       state1,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.GetVariable(headerKey).(int) < 3
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, ctx)

	// cycle one
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state1, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))

	// cycle two
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state1, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))

	// cycle three
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.state)
	assert.Equal(t, SMInProgress, sm.context.GetVariable(SMStatusHeaderKey))

	// cycle four
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.state)
	assert.Equal(t, SMEnd, sm.context.GetVariable(SMStatusHeaderKey))

}

func TestComplexCondition(t *testing.T) {
	state1 := &State{
		name: "state_1",
		action: func(ctx *context) error {
			log.Print("State 1 action")
			return nil
		},
	}

	state2 := &State{
		name: "state_2",
		action: func(ctx *context) error {
			log.Print("State 2 action")
			return nil
		},
	}

	state3 := &State{
		name: "state_3",
		action: func(ctx *context) error {
			log.Print("State 3 action")
			return nil
		},
	}

	state4 := &State{
		name: "state_4",
		action: func(ctx *context) error {
			log.Print("State 4 action")
			return nil
		},
	}

	state5 := &State{
		name: "state_5",
		action: func(ctx *context) error {
			log.Print("State 5 action")
			return nil
		},
	}

	state1.transitions = append(state1.transitions, Transition{
		to:       state2,
		priority: 2,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state1.transitions = append(state1.transitions, Transition{
		to:       state3,
		priority: 2,
		shouldTakeTransition: func(ctx context) bool {
			return false
		},
	})

	state2.transitions = append(state2.transitions, Transition{
		to:       state4,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state4.transitions = append(state4.transitions, Transition{
		to:       state5,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state3.transitions = append(state3.transitions, Transition{
		to:       state5,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state1.transitions = append(state1.transitions, Transition{
		to:       state5,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, ctx)

	for sm.hastNext() {
		_ = sm.next()
	}
}
