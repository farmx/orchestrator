package orchestrator


type TransactionalStep interface {
	DoAction(ctx *context) error
	UndoAction(ctx context)
}

type transactionStatus string

const (
	Condition     int = 2
	Default       int = 1

	Started transactionStatus = "STARTED"
	Success transactionStatus = "SUCCESS"
	Fail    transactionStatus = "FAIL"
)

type (
	route struct {
		// route identifier
		id             string

		rootStates *state
		lastState *state

		endFlag bool

		// statemachine ...
		statemachine *statemachine

		// conditionStateStack keep condition steps for otherwise/end-condition purpose
		predicateStack stateStack

		// transaction status
		status         transactionStatus

	}

	stateStack struct {
		stack []predicateState
	}

	predicateState struct {
		predicate func (context) bool
		state *state
	}
)

func (tss *stateStack) isEmpty() bool {
	return len(tss.stack) < 1
}

func (tss *stateStack) push(predicate func (context) bool, state *state) {
	tss.stack = append(tss.stack, predicateState{
		predicate: predicate,
		state: state,
	})
}

func (tss *stateStack) getLast() predicateState {
	stackLen := len(tss.stack)

	return tss.stack[stackLen - 1]
}

func (tss *stateStack) pop() predicateState {
	stackLen := len(tss.stack)

	s := tss.stack[stackLen - 1]
	tss.stack = tss.stack[:stackLen - 1]

	return s
}

// NewRoute define and return a route
func NewRoute(routeId string) *route {
	return &route{
		id:          routeId,
		statemachine: &statemachine{},
	}
}

// AddNextStep TODO: define retry on each state as a transition
// AddNextStep add new step to route
func (r *route) AddNextStep(step TransactionalStep) {
	s := &state{
		action: r.defineAction(step),
	}

	if r.rootStates == nil {
		r.rootStates = s
	}

	if r.endFlag {
		r.lastState = r.predicateStack.pop().state
		r.endFlag = false
	}

	if r.lastState != nil {
		r.defineTwoWayTransition(r.lastState, Default, func(ctx context) bool {
			return ctx.getVariable(SMStatusHeaderKey) != SMRollback
		}, s)
	}

	// update last state
	r.lastState = s
}

// When to define a condition
func (r *route) When(predicate func(ctx context) bool, step TransactionalStep) *route {
	s := &state{
		action: r.defineAction(step),
	}

	r.defineTwoWayTransition(r.lastState, Condition, predicate, s)
	r.predicateStack.push(predicate, s)

	// update last state
	r.lastState = s
	return r
}

// Otherwise when condition
func (r *route) Otherwise(step TransactionalStep) *route {
	s := &state{
		action: r.defineAction(step),
	}

	pr := r.predicateStack.getLast().predicate
	r.defineTwoWayTransition(r.lastState, Condition, func(ctx context) bool {
		return !pr(ctx)
	}, s)

	return r
}

// End of condition
func (r *route) End() *route {
	if r.endFlag {
		r.predicateStack.pop()
	}

	r.endFlag = true

	return r
}


func (r *route) Exec(ctx context) {
	r.status = Started
	r.statemachine.init(r.rootStates, ctx)

	for r.statemachine.hastNext() {
		_ = r.statemachine.next()
	}
}

func (r *route) defineAction(step TransactionalStep) func(ctx *context) error {
	return func(ctx *context) error {
		if ctx.getVariable(SMStatusHeaderKey) == SMRollback {
			step.UndoAction(*ctx)
			return nil
		}

		return step.DoAction(ctx)
	}
}

func (r *route) defineTwoWayTransition(src *state, priority int, predicate func (context) bool, dst *state) {
	// make a transition form prev state to last state
	src.transitions = append(src.transitions, transition{
		to: dst,
		priority: priority,
		shouldTakeTransition: predicate,
	})

	// make a transition from last to prev state on rollback
	dst.transitions = append(dst.transitions, transition{
		to: src,
		priority: Default,
		shouldTakeTransition: func(ctx context) bool {
			return ctx.getVariable(SMStatusHeaderKey) == SMRollback
		},
	})
}
