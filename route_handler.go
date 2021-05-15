package orchestrator

// TODO: define retry on each state as a transition

type routeStatus string

const (
	InProgress routeStatus = "IN_PROGRESS"
	Success    routeStatus = "SUCCESS"
	Fail       routeStatus = "FAIL"
)

type routeHandler struct {
	// transactionalRoute handler id
	id string

	// action transactionalRoute root state
	routeRootState *state

	// recovery transactionalRoute root state
	recoveryRootState *state

	// statemachine ...
	statemachine *statemachine

	// transactionalRoute transaction execution status
	status routeStatus
}

// transactionalRoute is root state
func newRouteHandler(routeRootState *state, recoveryRootState *state) *routeHandler {
	return &routeHandler{
		routeRootState:    routeRootState,
		recoveryRootState: recoveryRootState,
		statemachine:      &statemachine{},
	}
}

func (rh *routeHandler) exec(ctx *context, errCh chan<- error) {
	rh.statemachine.init(rh.routeRootState, ctx)
	rh.status = InProgress

	for rh.statemachine.hastNext() {
		err := rh.statemachine.next()
		mst, mctx := rh.statemachine.getMemento()

		if err == nil {
			continue
		}

		rh.status = Fail
		errCh <- err

		if rh.recoveryRootState != nil {
			rh.statemachine.init(rh.recoveryRootState, &mctx)

			// skip recovery transactionalRoute error
			for rh.statemachine.hastNext() {
				_ = rh.statemachine.next()
			}

			rh.statemachine.init(mst, &mctx)
		}
	}

	if rh.status != Fail {
		rh.status = Success
	}
}
