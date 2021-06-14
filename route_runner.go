package orchestrator

type routeStatus string

const (
	InProgress routeStatus = "IN_PROGRESS"
	Done       routeStatus = "DONE"
)

type routeRunner struct {
	// TransactionalRoute handler id
	id string

	// action TransactionalRoute root State
	routeRootState *State

	// recovery TransactionalRoute root State
	recoveryRootState *State

	// statemachine ...
	statemachine *statemachine

	// TransactionalRoute transaction execution status
	status routeStatus
}

func newRouteRunner(routeRootState *State, recoveryRootState *State) *routeRunner {
	return &routeRunner{
		routeRootState:    routeRootState,
		recoveryRootState: recoveryRootState,
		statemachine:      &statemachine{},
	}
}

func (rr *routeRunner) run(ctx *context, errCh chan<- error) {
	rr.statemachine.init(rr.routeRootState, ctx)
	rr.status = InProgress

	for rr.statemachine.hastNext() {
		err := rr.statemachine.next()
		mst, mctx := rr.statemachine.getMemento()

		if err == nil {
			continue
		}

		errCh <- err

		if rr.recoveryRootState != nil {
			rr.statemachine.init(rr.recoveryRootState, &mctx)

			for rr.statemachine.hastNext() {
				errCh <- rr.statemachine.next()
			}

			rr.statemachine.init(mst, &mctx)
		}
	}

	rr.status = Done
}

func (rr *routeRunner) shutdown() {

}
