package orchestrator

// TODO: define retry on each State as a transition

type routeStatus string

const (
	InProgress routeStatus = "IN_PROGRESS"
	Success    routeStatus = "SUCCESS"
	Fail       routeStatus = "FAIL"
)

type routeRunner struct {
	// TransactionalRoute handler id
	id string

	// action TransactionalRoute root state
	routeRootState *state

	// recovery TransactionalRoute root state
	recoveryRootState *state

	// statemachine ...
	statemachine *statemachine

	// TransactionalRoute transaction execution status
	status routeStatus
}

func newRouteRunner(routeRootState *state, recoveryRootState *state) *routeRunner {
	return &routeRunner{
		routeRootState:    routeRootState,
		recoveryRootState: recoveryRootState,
		statemachine:      &statemachine{},
	}
}

func (rr *routeRunner) exec(ctx *context, errCh chan<- error) {
	rr.statemachine.init(rr.routeRootState, ctx)
	rr.status = InProgress

	for rr.statemachine.hastNext() {
		err := rr.statemachine.next()
		mst, mctx := rr.statemachine.getMemento()

		if err == nil {
			continue
		}

		rr.status = Fail
		errCh <- err

		if rr.recoveryRootState != nil {
			rr.statemachine.init(rr.recoveryRootState, &mctx)

			// skip recovery TransactionalRoute error
			for rr.statemachine.hastNext() {
				_ = rr.statemachine.next()
			}

			rr.statemachine.init(mst, &mctx)
		}
	}

	if rr.status != Fail {
		rr.status = Success
	}
}
