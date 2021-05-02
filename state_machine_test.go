package orchestrator

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestHappyScenario(t *testing.T) {
	state1 := &state{
		action: func(ctx *context) error {
			log.Print("state 1 action")
			return nil
		},
	}

	state2 := &state{
		action: func(ctx *context) error {
			log.Print("state 2 action")
			return nil
		},
	}

	state1.transitions = append(state1.transitions, transition{
		to: state2,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, *ctx)

	// init time
	assert.Equal(t, true, sm.hastNext())
	assert.Equal(t, state1, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))

	// cycle one
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))
	assert.Equal(t, true, sm.hastNext())

	// cycle three
	assert.Nil(t, sm.next())
	assert.Equal(t, false, sm.hastNext())
	assert.Equal(t, state2, sm.currentState)
	assert.Equal(t, SMEnd, sm.context.getVariable(SMStatusHeaderKey))
}

func TestRollback(t *testing.T) {
	state1 := &state{
		action: func(ctx *context) error {
			if ctx.getVariable(SMStatusHeaderKey) == SMRollback {
				log.Print("rollback call")
				return nil
			}

			log.Print("state 1 action")
			return nil
		},
	}

	state2 := &state{
		action: func(ctx *context) error {
			log.Print("state 2 action")
			return errors.New("fake error")
		},
	}

	state1.transitions = append(state1.transitions, transition{
		to: state2,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.getVariable(SMStatusHeaderKey) != SMRollback
		},
	})

	state2.transitions = append(state2.transitions, transition{
		to: state1,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.getVariable(SMStatusHeaderKey) == SMRollback
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, *ctx)


	// init time
	assert.Equal(t, true, sm.hastNext())
	assert.Equal(t, state1, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))

	// cycle one
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))
	assert.Equal(t, true, sm.hastNext())

	// cycle two
	assert.NotNil(t, sm.next())
	assert.Equal(t, state1, sm.currentState)
	assert.Equal(t, SMRollback, sm.context.getVariable(SMStatusHeaderKey))
	assert.Equal(t, true, sm.hastNext())

	// cycle three
	assert.Nil(t, sm.next())
	assert.Equal(t, false, sm.hastNext())
	assert.Equal(t, state1, sm.currentState)
	assert.Equal(t, SMEnd, sm.context.getVariable(SMStatusHeaderKey))
}

func TestLoop(t *testing.T) {
	headerKey := "HEADER_KEY"

	state1 := &state{
		action: func(ctx *context) error {
			if v := ctx.getVariable(headerKey); v == nil {
				ctx.setVariable(headerKey, 0)
			}

			ctx.setVariable(headerKey, ctx.getVariable(headerKey).(int) + 1)
			log.Print("state 1 action")
			return nil
		},
	}

	state2 := &state{
		action: func(ctx *context) error {
			log.Print("state 2 action")
			return nil
		},
	}

	state1.transitions = append(state1.transitions, transition{
		to: state2,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.getVariable(headerKey) == 3
		},
	})

	state1.transitions = append(state1.transitions, transition{
		to: state1,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.getVariable(headerKey).(int) < 3
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, *ctx)

	// cycle one
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state1, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))

	// cycle two
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state1, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))

	// cycle three
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.currentState)
	assert.Equal(t, SMInProgress, sm.context.getVariable(SMStatusHeaderKey))

	// cycle four
	assert.Equal(t, true, sm.hastNext())
	assert.Nil(t, sm.next())
	assert.Equal(t, state2, sm.currentState)
	assert.Equal(t, SMEnd, sm.context.getVariable(SMStatusHeaderKey))

}

func TestComplexCondition(t *testing.T) {
	state1 := &state{
		name: "state_1",
		action: func(ctx *context) error {
			log.Print("state 1 action")
			return nil
		},
	}

	state2 := &state{
		name: "state_2",
		action: func(ctx *context) error {
			log.Print("state 2 action")
			return nil
		},
	}

	state3 := &state{
		name: "state_3",
		action: func(ctx *context) error {
			log.Print("state 3 action")
			return nil
		},
	}

	state4 := &state{
		name: "state_4",
		action: func(ctx *context) error {
			log.Print("state 4 action")
			return nil
		},
	}

	state5 := &state{
		name: "state_5",
		action: func(ctx *context) error {
			log.Print("state 5 action")
			return nil
		},
	}

	state1.transitions = append(state1.transitions, transition{
		to: state2,
		priority: 2,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state1.transitions = append(state1.transitions, transition{
		to: state3,
		priority: 2,
		shouldTakeTransition: func(ctx context) bool {
			return false
		},
	})

	state2.transitions = append(state2.transitions, transition{
		to: state4,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state4.transitions = append(state4.transitions, transition{
		to: state5,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state3.transitions = append(state3.transitions, transition{
		to: state5,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	state1.transitions = append(state1.transitions, transition{
		to: state5,
		priority: 1,
		shouldTakeTransition: func(ctx context) bool {
			return true
		},
	})

	ctx, _ := NewContext()
	sm := &statemachine{}
	sm.init(state1, *ctx)

	for sm.hastNext() {
		_ = sm.next()
	}
}