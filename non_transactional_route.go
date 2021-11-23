package orchestrator

import "fmt"

type (
	NonTransactionalRoute struct {
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
	onlyNonTRAddNextStep interface {
		AddNextStep(name string, doAction func(ctx *context) error) *NonTransactionalRoute
	}
)

// NewNonTransactionalRoute define and return a NonTransactionalRoute
func NewNonTransactionalRoute(id string) *NonTransactionalRoute {
	return &NonTransactionalRoute{
		id:         id,
		routeState: Main,
	}
}

// AddNextStep add new step to NonTransactionalRoute
func (ntr *NonTransactionalRoute) AddNextStep(name string, doAction func(ctx *context) error) *NonTransactionalRoute {
	s := &State{
		name:   fmt.Sprintf("%s_%s", ntr.id, name),
		action: doAction,
	}

	switch ntr.routeState {
	case When:
		ntr.addNextStepAfterWhen(s)
		break
	case Else:
		ntr.addNextStepAfterOtherwise(s)
		break
	case End:
		ntr.addNextStepAfterEnd(s)
		break
	default:
		// first state must be define as a start start (root)
		if ntr.startState == nil {
			ntr.startState = s
			break
		}

		ntr.lastState.createTransition(s, Default, func(ctx context) bool {
			return true
		})
	}

	// update last State
	ntr.lastState = s
	ntr.routeState = Main
	return ntr
}

func (ntr *NonTransactionalRoute) addNextStepAfterWhen(s *State) {
	ntr.lastState.createTransition(s, Condition, ntr.predicateStateStack.getLast().predicate)
}

func (ntr *NonTransactionalRoute) addNextStepAfterOtherwise(s *State) {
	ps := ntr.predicateStateStack.getLast()

	ps.state.createTransition(s, Condition, func(ctx context) bool {
		return !ps.predicate(ctx)
	})
}

//        condition       condition
//       /         \        |   \
//     not         yes      no   yes
//  included        |       |    |
//       \         /        |   /
//        End State       End State
func (ntr *NonTransactionalRoute) addNextStepAfterEnd(s *State) {
	cs := ntr.predicateStateStack.pop().state
	predicate := func(ctx context) bool {
		return true
	}

	// define Transition from last State of each condition State
	states := ntr.getEachTransitionLatestState(cs)
	for _, es := range states {
		es.createTransition(s, Default, predicate)
	}

	// Otherwise doesn't define
	if len(states) < 2 {
		// define a Transition from latest state with a conditional transition
		cs.createTransition(s, Default, predicate)
	}
}

// When to define a condition
func (ntr *NonTransactionalRoute) When(predicate func(ctx context) bool) onlyNonTRAddNextStep {
	ntr.routeState = When
	ntr.predicateStateStack.push(predicate, ntr.lastState)

	return ntr
}

// Otherwise When condition
func (ntr *NonTransactionalRoute) Otherwise() onlyNonTRAddNextStep {
	ntr.routeState = Else

	return ntr
}

// End of condition
func (ntr *NonTransactionalRoute) End() onlyNonTRAddNextStep {
	ntr.routeState = End

	return ntr
}

func (ntr *NonTransactionalRoute) To(id string) *NonTransactionalRoute {
	ntr.endpoints = append(ntr.endpoints, &Endpoint{
		To:    id,
		State: ntr.lastState,
	})

	return ntr
}

func (ntr *NonTransactionalRoute) GetRouteId() string {
	return ntr.id
}

func (ntr *NonTransactionalRoute) GetStartState() *State {
	return ntr.startState
}

func (ntr *NonTransactionalRoute) GetEndpoints() []*Endpoint {
	return ntr.endpoints
}

func (ntr *NonTransactionalRoute) getEachTransitionLatestState(state *State) []*State {
	var result []*State
	for _, t := range state.transitions {
		if t.priority == Condition {
			result = append(result, ntr.getLatestState(t.to))
		}
	}

	return result
}

// looking for latest state
func (ntr *NonTransactionalRoute) getLatestState(state *State) *State {
	for _, st := range state.transitions {
		if st.priority == Default {
			return ntr.getLatestState(st.to)
		}
	}

	return state
}
