package orchestrator

const (
	Condition int = 2
	Default   int = 1
)

type (
	route struct {

		// rootStates graph root state
		rootStates *state

		// latest added state
		lastState *state

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

// newRoute define and return a route
func newRoute() *route {
	return &route{
		counter: newCounter(),
	}
}

// addNextStep add new step to route
func (r *route) addNextStep(doAction func(ctx *context) error,undoAction func(ctx context)) *route {
	s := &state{
		name:   "state_" + r.counter.next(),
		action: r.defineAction(doAction, undoAction),
	}

	if r.rootStates == nil {
		r.rootStates = s
	}

	if r.lastState != nil {
		r.defineTwoWayTransition(r.lastState, Default, func(ctx context) bool {
			return ctx.GetVariable(SMStatusHeaderKey) != SMRollback
		}, s)
	}

	// update last state
	r.lastState = s
	return r
}

// when to define a condition
func (r *route) when(predicate func(ctx context) bool, doAction func(ctx *context) error,undoAction func(ctx context)) *route {
	s := &state{
		name:   "state_c_" + r.counter.subCount(),
		action: r.defineAction(doAction, undoAction),
	}

	r.predicateStack.push(predicate, r.lastState)
	r.defineTwoWayTransition(r.lastState, Condition, predicate, s)

	// update last state
	r.lastState = s
	return r
}

// otherwise when condition
func (r *route) otherwise(doAction func(ctx *context) error,undoAction func(ctx context)) *route {
	r.counter.endSubCounting()
	s := &state{
		name:   "state_!c_" + r.counter.subCount(),
		action: r.defineAction(doAction, undoAction),
	}

	ps := r.predicateStack.getLast()
	r.defineTwoWayTransition(ps.state, Condition, func(ctx context) bool {
		return !ps.predicate(ctx)
	}, s)

	r.lastState = s
	return r
}

//        condition       condition
//       /         \        |   \
//     not         yes      no   yes
//  included        |       |    |
//       \         /        |   /
//        end state       end state

// end of condition
func (r *route) end(doAction func(ctx *context) error,undoAction func(ctx context)) *route {
	r.counter.endSubCounting()

	s := &state{
		name:   "state_" + r.counter.next(),
		action: r.defineAction(doAction, undoAction),
	}

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

	r.lastState = s
	return r
}

func (r *route) getRouteStateMachine() *state {
	return r.rootStates
}

func (r *route) defineAction(doAction func(ctx *context) error,undoAction func(ctx context)) func(ctx *context) error {
	return func(ctx *context) error {
		if ctx.GetVariable(SMStatusHeaderKey) == SMRollback {
			undoAction(*ctx)
			return nil
		}

		return doAction(ctx)
	}
}

func (r *route) defineTwoWayTransition(src *state, priority int, predicate func(context) bool, dst *state) {
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

func (r *route) getConditionalLastStates(root *state) []*state {
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
