package orchestrator

import "sort"

type statemachineStatus string

const (
	SMStatusHeaderKey string = "SM_STATUS"

	SMInProgress statemachineStatus = "IN_PROGRESS"
	SMRollback   statemachineStatus = "ROLLBACK"
	SMEnd        statemachineStatus = "END"
)

type (
	statemachine struct {
		currentState *State
		context      *context
	}

	State struct {
		name        string
		transitions []Transition
		action      func(ctx *context) error
	}

	Transition struct {
		to                   *State
		priority             int
		shouldTakeTransition func(ctx context) bool
	}
)

func (sm *statemachine) init(currentState *State, ctx *context) {
	sm.currentState = currentState
	sm.context = ctx

	if sm.context.GetVariable(SMStatusHeaderKey) == nil {
		sm.context.SetVariable(SMStatusHeaderKey, SMInProgress)
	}
}

func (sm *statemachine) hastNext() bool {
	return sm.context.GetVariable(SMStatusHeaderKey) != SMEnd
}

func (sm *statemachine) next() (err error) {
	if err = sm.currentState.action(sm.context); err != nil {
		sm.context.SetVariable(SMStatusHeaderKey, SMRollback)
	}

	// sort based on priority
	sort.Slice(sm.currentState.transitions[:], func(i, j int) bool {
		return sm.currentState.transitions[i].priority >= sm.currentState.transitions[j].priority
	})

	for _, ts := range sm.currentState.transitions {
		if ts.shouldTakeTransition(*sm.context) {
			sm.currentState = ts.to
			return err
		}
	}

	sm.context.SetVariable(SMStatusHeaderKey, SMEnd)
	return err
}

func (sm *statemachine) getMemento() (*State, context) {
	return sm.currentState, *sm.context
}
