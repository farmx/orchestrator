package orchestrator

type (
	predicateStateStack struct {
		stack []predicateState
	}

	predicateState struct {
		predicate func(context) bool
		state     *State
	}
)

func (tss *predicateStateStack) isEmpty() bool {
	return len(tss.stack) < 1
}

func (tss *predicateStateStack) push(predicate func(context) bool, state *State) {
	tss.stack = append(tss.stack, predicateState{
		predicate: predicate,
		state:     state,
	})
}

func (tss *predicateStateStack) getLast() predicateState {
	stackLen := len(tss.stack)

	return tss.stack[stackLen-1]
}

func (tss *predicateStateStack) pop() predicateState {
	stackLen := len(tss.stack)

	s := tss.stack[stackLen-1]
	tss.stack = tss.stack[:stackLen-1]

	return s
}
