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
	transactionalRoute struct {

		// rootStates graph root state
		rootStates *state

		// route state
		state transitionState

		// latest added state
		lastState *state

		// state name prefix
		statePrefix string

		// conditionStateStack keep condition steps for otherwise/end-condition purpose
		predicateStack stateStack

		// naming the state
		counter *counter
	}

	stateStack struct {
		stack []predicateState
	}

	predicateState struct {
		predicate func(context) bool
		state     *state
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

// newRoute define and return a transactionalRoute
func newRoute() *transactionalRoute {
	return &transactionalRoute{
		counter:     newCounter(),
		state:       Main,
		statePrefix: DefaultPrefix,
	}
}

// addNextStep add new step to transactionalRoute
func (r *transactionalRoute) addNextStep(doAction func(ctx *context) error, undoAction func(ctx context)) *transactionalRoute {
	s := &state{
		name:   r.statePrefix + r.counter.next(),
		action: r.defineAction(doAction, undoAction),
	}

	switch r.state {
	case When:
		r.addNextStepAfterWhen(s)
		break
	case Else:
		r.addNextStepAfterOtherwise(s)
		break
	case End:
		r.addNextStepAfterEnd(s)
		break
	default:
		if r.rootStates == nil {
			r.rootStates = s
			break
		}

		r.defineTwoWayTransition(r.lastState, Default, func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) != SMRollback
		}, s)
	}

	// update last state
	r.lastState = s
	r.state = Main
	return r
}

func (r *transactionalRoute) addNextStepAfterWhen(s *state) {
	r.defineTwoWayTransition(r.lastState, Condition, r.predicateStack.getLast().predicate, s)
}

func (r *transactionalRoute) addNextStepAfterOtherwise(s *state) {
	ps := r.predicateStack.getLast()

	r.defineTwoWayTransition(ps.state, Condition, func(ctx context) bool {
		return !ps.predicate(ctx)
	}, s)
}

//        condition       condition
//       /         \        |   \
//     not         yes      no   yes
//  included        |       |    |
//       \         /        |   /
//        end state       end state

func (r *transactionalRoute) addNextStepAfterEnd(s *state) {
	predicate := func(ctx context) bool {
		return true
	}

	cs := r.predicateStack.pop().state

	// define transition from last state of each condition state
	cls := r.getConditionalLastStates(cs)
	for _, es := range cls {
		r.defineTwoWayTransition(es, Default, predicate, s)
	}

	// otherwise doesn't define
	if len(cls) < 2 {
		// define a transition from root condition state
		r.defineTwoWayTransition(cs, Default, predicate, s)
	}
}

// when to define a condition
func (r *transactionalRoute) when(predicate func(ctx context) bool) *transactionalRoute {
	r.state = When
	r.predicateStack.push(predicate, r.lastState)

	// state naming
	r.statePrefix = ConditionPrefix
	r.counter.subVersioning()

	return r
}

// otherwise when condition
func (r *transactionalRoute) otherwise() *transactionalRoute {
	r.state = Else

	// state naming
	r.statePrefix = OtherwisePrefix
	r.counter.endSubVersioning()
	r.counter.subVersioning()

	return r
}

// end of condition
func (r *transactionalRoute) end() *transactionalRoute {
	r.state = End

	// state naming
	r.statePrefix = DefaultPrefix
	r.counter.endSubVersioning()

	return r
}

func (r *transactionalRoute) getRouteStateMachine() *state {
	return r.rootStates
}

func (r *transactionalRoute) defineAction(doAction func(ctx *context) error, undoAction func(ctx context)) func(ctx *context) error {
	return func(ctx *context) error {
		if ctx.GetVariable(SMStatusHeaderKey) == SMRollback {
			undoAction(*ctx)
			return nil
		}

		return doAction(ctx)
	}
}

func (r *transactionalRoute) defineTwoWayTransition(src *state, priority int, predicate func(context) bool, dst *state) {
	// define a transition form src state to dst state
	src.transitions = append(src.transitions, transition{
		to:                   dst,
		priority:             priority,
		shouldTakeTransition: predicate,
	})

	// define a transition from dst to src state for rollback
	dst.transitions = append(dst.transitions, transition{
		to:       src,
		priority: Default,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) == SMRollback
		},
	})
}

func (r *transactionalRoute) getConditionalLastStates(root *state) []*state {
	var result []*state
	for _, tr := range root.transitions {
		if tr.priority == Condition {
			result = append(result, lastState(tr.to))
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
