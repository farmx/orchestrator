package orchestrator

import "fmt"

const (
	transactionStatusHeaderKey  = "TRANSACTION_STATUS"
	transactionStatusRollback   = "ROLLBACK"
	transactionStatusInProgress = "IN_PROGRESS"
	transactionStatusEnd        = "END"
)

type (
	TransactionalRoute struct {
		// route id
		id string

		// startState graph root state
		startState *State

		// route state
		routeState routeState

		// latest added state
		lastState *State

		// predicateStateStack keep conditional state for Otherwise/End-condition transition
		predicateStateStack predicateStateStack

		// endpoint list
		endpoints []*Endpoint
	}

	// force to present AddNextStep method only
	onlyTRAddNextStep interface {
		AddNextStep(name string, doAction func(ctx *context) error, undoAction func(ctx context)) *TransactionalRoute
	}
)

// NewTransactionalRoute define and return a TransactionalRoute
func NewTransactionalRoute(id string) *TransactionalRoute {
	return &TransactionalRoute{
		id:             id,
		routeState:          Main,
	}
}

// AddNextStep add new step to TransactionalRoute
func (tr *TransactionalRoute) AddNextStep(name string, doAction func(ctx *context) error, undoAction func(ctx context)) *TransactionalRoute {
	s := &State{
		name:   fmt.Sprintf("%s_%s", tr.id, name),
		action: tr.defineAction(doAction, undoAction),
	}

	switch tr.routeState {
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
			return ctx.GetVariable(transactionStatusHeaderKey) != transactionStatusRollback
		}, s)
	}

	// update last State
	tr.lastState = s
	tr.routeState = Main
	return tr
}

func (tr *TransactionalRoute) addNextStepAfterWhen(s *State) {
	tr.defineTwoWayTransition(tr.lastState, Condition, tr.predicateStateStack.getLast().predicate, s)
}

func (tr *TransactionalRoute) addNextStepAfterOtherwise(s *State) {
	ps := tr.predicateStateStack.getLast()

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
func (tr *TransactionalRoute) addNextStepAfterEnd(s *State) {
	cs := tr.predicateStateStack.pop().state
	predicate := func(ctx context) bool {
		return true
	}

	// define Transition from last State of each condition State
	states := tr.getEachTransitionLatestState(cs)
	for _, es := range states {
		tr.defineTwoWayTransition(es, Default, predicate, s)
	}

	// Otherwise doesn't define
	if len(states) < 2 {
		// define a Transition from latest state with a conditional transition
		tr.defineTwoWayTransition(cs, Default, predicate, s)
	}
}

// When to define a condition
func (tr *TransactionalRoute) When(predicate func(ctx context) bool) onlyTRAddNextStep {
	tr.routeState = When
	tr.predicateStateStack.push(predicate, tr.lastState)

	return tr
}

// Otherwise When condition
func (tr *TransactionalRoute) Otherwise() onlyTRAddNextStep {
	tr.routeState = Else

	return tr
}

// End of condition
func (tr *TransactionalRoute) End() onlyTRAddNextStep {
	tr.routeState = End

	return tr
}

func (tr *TransactionalRoute) To(id string) *TransactionalRoute {
	tr.endpoints = append(tr.endpoints, &Endpoint{
		To:    id,
		State: tr.lastState,
	})

	return tr
}

func (tr *TransactionalRoute) GetRouteId() string {
	return tr.id
}

func (tr *TransactionalRoute) GetStartState() *State {
	return tr.startState
}

func (tr *TransactionalRoute) GetEndpoints() []*Endpoint {
	return tr.endpoints
}

func (tr *TransactionalRoute) defineAction(doAction func(ctx *context) error, undoAction func(ctx context)) func(ctx *context) error {
	return func(ctx *context) error {
		if ctx.GetVariable(transactionStatusHeaderKey) == transactionStatusRollback {
			undoAction(*ctx)
			return nil
		}

		return doAction(ctx)
	}
}

func (tr *TransactionalRoute) defineTwoWayTransition(src *State, priority int, predicate func(context) bool, dst *State) {
	// define a Transition form src State to dst State
	src.createTransition(dst, priority,
		 func(ctx context) bool {
			return predicate(ctx) && ctx.GetVariable(transactionStatusHeaderKey) != transactionStatusRollback
		})

	// define a Transition from dst to src State for rollback
	dst.createTransition(src, Default,
		func(ctx context) bool {
			return ctx.GetVariable(transactionStatusHeaderKey) == transactionStatusRollback
		})
}

func (tr *TransactionalRoute) getEachTransitionLatestState(state *State) []*State {
	var result []*State
	for _, t := range state.transitions {
		if t.priority == Condition {
			result = append(result, getLatestState(t.to))
		}
	}

	return result
}

// looking for latest state
func getLatestState(state *State) *State {
	for _, tr := range state.transitions {
		ctx, _ := NewContext()
		ctx.SetVariable(transactionStatusHeaderKey, transactionStatusRollback)
		// choose happy path transition
		if tr.priority == Default && !tr.shouldTakeTransition(*ctx) {
			return getLatestState(tr.to)
		}
	}

	return state
}
