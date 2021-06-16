package orchestrator

import (
	"sort"
	"time"
)

type statemachineStatus string

const (
	SMStatusHeaderKey string = "SM_STATUS"

	SMInProgress statemachineStatus = "IN_PROGRESS"
	SMEnd        statemachineStatus = "END"
)

type (
	statemachine struct {
		state   *State
		context *context
	}

	State struct {
		name          string
		transitions   []Transition
		action        func(ctx *context) error
		actionTimeout time.Duration
	}

	Transition struct {
		to                   *State
		priority             int
		shouldTakeTransition func(ctx context) bool
	}
)

func (sm *statemachine) init(state *State, ctx *context) {
	sm.state = state
	sm.context = ctx

	if sm.context.GetVariable(SMStatusHeaderKey) == nil {
		sm.context.SetVariable(SMStatusHeaderKey, SMInProgress)
	}
}

func (sm *statemachine) hastNext() bool {
	return sm.context.GetVariable(SMStatusHeaderKey) != SMEnd
}

func (sm *statemachine) next() (err error) {
	err = sm.state.action(sm.context)

	// sort based on priority
	sort.Slice(sm.state.transitions[:], func(i, j int) bool {
		return sm.state.transitions[i].priority >= sm.state.transitions[j].priority
	})

	for _, ts := range sm.state.transitions {
		if ts.shouldTakeTransition(*sm.context) {
			sm.state = ts.to
			return err
		}
	}

	sm.context.SetVariable(SMStatusHeaderKey, SMEnd)
	return err
}

func (sm *statemachine) getMemento() (*State, context) {
	return sm.state, *sm.context
}
