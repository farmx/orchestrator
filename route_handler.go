package orchestrator

// TODO: define retry on each state as a transition

type transactionStatus string

const (
	InProgress transactionStatus = "IN_PROGRESS"
	Success    transactionStatus = "SUCCESS"
	Fail       transactionStatus = "FAIL"
)

type routeHandler struct {
	// route handler id
	id string

	// action route root state
	routeRootState *state

	// recovery route root state
	recoveryRootState *state

	// statemachine ...
	statemachine *statemachine

	// route transaction execution status
	status transactionStatus
}

// route is root state
func newRouteHandler(id string, routeRootState *state, recoveryRootState *state) *routeHandler {
	return &routeHandler{
		id:                id,
		routeRootState:    routeRootState,
		recoveryRootState: recoveryRootState,
		statemachine:      &statemachine{},
	}
}

func (rh *routeHandler) exec(ctx context) {
	rh.statemachine.init(rh.routeRootState, ctx)
	rh.status = InProgress

	for rh.statemachine.hastNext() {
		err := rh.statemachine.next()
		mst, mctx := rh.statemachine.getMemento()

		if err == nil {
			continue
		}

		rh.status = Fail

		if rh.recoveryRootState != nil {
			rh.statemachine.init(rh.recoveryRootState, mctx)

			// skip recovery route error
			for rh.statemachine.hastNext() {
				_ = rh.statemachine.next()
			}

			rh.statemachine.init(mst, mctx)
		}
	}

	if rh.status != Fail {
		rh.status = Success
	}
}
