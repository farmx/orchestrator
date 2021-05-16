package orchestrator

type transitionState string

const (
	Main transitionState = "MAIN"
	When transitionState = "WHEN"
	Else transitionState = "ELSE"
	End  transitionState = "END"

	Condition int = 2
	Default   int = 1

	DefaultPrefix   = "s_"
	ConditionPrefix = "sc_"
	OtherwisePrefix = "sc!_"
)

type (
	TransactionalRoute struct {

		// startState graph root state
		startState *state

		// route state
		state transitionState

		// latest added state
		lastState *state

		// state name prefix
		statePrefix string

		// conditionStateStack keep condition steps for Otherwise/End-condition purpose
		predicateStack stateStack

		// naming the state
		counter *counter

		// endpoint list
		endpoints []*Endpoint
	}

	stateStack struct {
		stack []predicateState
	}

	predicateState struct {
		predicate func(context) bool
		state     *state
	}

	onlyProcessor interface {
		AddNextStep(doAction func(ctx *context) error, undoAction func(ctx context)) *TransactionalRoute
	}
)

func (tss *stateStack) isEmpty() bool {
	return len(tss.stack) < 1
}

func (tss *stateStack) push(predicate func(context) bool, state *state) {
	tss.stack = append(tss.stack, predicateState{
		predicate: predicate,
		state:     state,
	})
}

func (tss *stateStack) getLast() predicateState {
	stackLen := len(tss.stack)

	return tss.stack[stackLen-1]
}

func (tss *stateStack) pop() predicateState {
	stackLen := len(tss.stack)

	s := tss.stack[stackLen-1]
	tss.stack = tss.stack[:stackLen-1]

	return s
}

// newTransactionalRoute define and return a TransactionalRoute
func newTransactionalRoute() *TransactionalRoute {
	return &TransactionalRoute{
		counter:     newCounter(),
		state:       Main,
		statePrefix: DefaultPrefix,
	}
}

// AddNextStep add new step to TransactionalRoute
func (tr *TransactionalRoute) AddNextStep(doAction func(ctx *context) error, undoAction func(ctx context)) *TransactionalRoute {
	s := &state{
		name:   tr.statePrefix + tr.counter.next(),
		action: tr.defineAction(doAction, undoAction),
	}

	switch tr.state {
	case When:
		tr.addNextStepAfterWhen(s)
		break
	case Else:
		tr.addNextStepAfterOtherwise(s)
		break
	case End:
		tr.addNextStepAfterEnd(s)
		break
	default:
		if tr.startState == nil {
			tr.startState = s
			break
		}

		tr.defineTwoWayTransition(tr.lastState, Default, func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) != SMRollback
		}, s)
	}

	// update last State
	tr.lastState = s
	tr.state = Main
	return tr
}

func (tr *TransactionalRoute) addNextStepAfterWhen(s *state) {
	tr.defineTwoWayTransition(tr.lastState, Condition, tr.predicateStack.getLast().predicate, s)
}

func (tr *TransactionalRoute) addNextStepAfterOtherwise(s *state) {
	ps := tr.predicateStack.getLast()

	tr.defineTwoWayTransition(ps.state, Condition, func(ctx context) bool {
		return !ps.predicate(ctx)
	}, s)
}

//        condition       condition
//       /         \        |   \
//     not         yes      no   yes
//  included        |       |    |
//       \         /        |   /
//        End State       End State

func (tr *TransactionalRoute) addNextStepAfterEnd(s *state) {
	predicate := func(ctx context) bool {
		return true
	}

	cs := tr.predicateStack.pop().state

	// define transition from last State of each condition State
	cls := tr.getConditionalLastStates(cs)
	for _, es := range cls {
		tr.defineTwoWayTransition(es, Default, predicate, s)
	}

	// Otherwise doesn't define
	if len(cls) < 2 {
		// define a transition from root condition State
		tr.defineTwoWayTransition(cs, Default, predicate, s)
	}
}

// When to define a condition
func (tr *TransactionalRoute) When(predicate func(ctx context) bool) onlyProcessor {
	tr.state = When
	tr.predicateStack.push(predicate, tr.lastState)

	// State naming
	tr.statePrefix = ConditionPrefix
	tr.counter.subVersioning()

	return tr
}

// Otherwise When condition
func (tr *TransactionalRoute) Otherwise() onlyProcessor {
	tr.state = Else

	// State naming
	tr.statePrefix = OtherwisePrefix
	tr.counter.endSubVersioning()
	tr.counter.subVersioning()

	return tr
}

// End of condition
func (tr *TransactionalRoute) End() onlyProcessor {
	tr.state = End

	// State naming
	tr.statePrefix = DefaultPrefix
	tr.counter.endSubVersioning()

	return tr
}

func (tr *TransactionalRoute) To(id string) *TransactionalRoute {
	tr.endpoints = append(tr.endpoints, &Endpoint{
		To:    id,
		State: tr.lastState,
	})

	return tr
}

func (tr *TransactionalRoute) GetStartState() *state {
	return tr.startState
}

func (tr *TransactionalRoute) GetEndpoints() []*Endpoint {
	return tr.endpoints
}

func (tr *TransactionalRoute) defineAction(doAction func(ctx *context) error, undoAction func(ctx context)) func(ctx *context) error {
	return func(ctx *context) error {
		if ctx.GetVariable(SMStatusHeaderKey) == SMRollback {
			undoAction(*ctx)
			return nil
		}

		return doAction(ctx)
	}
}

func (tr *TransactionalRoute) defineTwoWayTransition(src *state, priority int, predicate func(context) bool, dst *state) {
	// define a transition form src State to dst State
	src.transitions = append(src.transitions, transition{
		to:                   dst,
		priority:             priority,
		shouldTakeTransition: predicate,
	})

	// define a transition from dst to src State for rollback
	dst.transitions = append(dst.transitions, transition{
		to:       src,
		priority: Default,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) == SMRollback
		},
	})
}

func (tr *TransactionalRoute) getConditionalLastStates(root *state) []*state {
	var result []*state
	for _, t := range root.transitions {
		if t.priority == Condition {
			result = append(result, lastState(t.to))
		}
	}

	return result
}

func lastState(state *state) *state {
	for _, tr := range state.transitions {
		ctx, _ := NewContext()
		ctx.SetVariable(SMStatusHeaderKey, SMRollback)
		if tr.priority == Default && !tr.shouldTakeTransition(*ctx) {
			return lastState(tr.to)
		}
	}

	return state
}
