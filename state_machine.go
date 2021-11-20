package orchestrator

import (
	"sort"
	"time"
)

type (
	statemachine struct {
		state              *State
		context            *context
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

func (sm *statemachine) init(startState *State, ctx *context) {
	sm.state = startState
	sm.context = ctx
}

func (sm *statemachine) doAction() (bool, error) {
	err := sm.state.action(sm.context)

	// TODO: <Decision making> the priority can be dynamic according to the context values or static and cache it for performance improvement
	// sort based on priority
	sort.Slice(sm.state.transitions[:], func(i, j int) bool {
		return sm.state.transitions[i].priority >= sm.state.transitions[j].priority
	})

	for _, ts := range sm.state.transitions {
		if ts.shouldTakeTransition(*sm.context) {
			sm.state = ts.to
			return true, err
		}
	}

	return false, err
}

func (sm *statemachine) getMemento() (*State, context) {
	return sm.state, *sm.context
}

func (s *State) createTransition(to *State, priority int, shouldTakeTransition func(ctx context) bool) {
	s.transitions = append(s.transitions, Transition{
		to:       to,
		priority: priority,
		shouldTakeTransition: shouldTakeTransition,
	})
}
