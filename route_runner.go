package orchestrator

type routeRunner struct {
	// runner id
	id string

	// route root State
	routeRootState *State

	// recovery route root State
	recoveryRootState *State

	// statemachine ...
	statemachine *statemachine
}

func newRouteRunner(routeRootState *State, recoveryRootState *State) *routeRunner {
	return &routeRunner{
		routeRootState:    routeRootState,
		recoveryRootState: recoveryRootState,
		statemachine:      &statemachine{},
	}
}

func (rr *routeRunner) run(ctx *context, errCh chan<- error) {
	var err error

	rr.statemachine.init(rr.routeRootState, ctx)

	for hasNext := true ;  hasNext == true; hasNext, err = rr.statemachine.doAction() {
		mst, mctx := rr.statemachine.getMemento()

		if errCh != nil && err != nil {
			errCh <- err
		}

		// call error recovery handler
		if err != nil && rr.recoveryRootState != nil {
			rr.statemachine.init(rr.recoveryRootState, &mctx)

			for rcHasNext := true ;  rcHasNext == true; rcHasNext, err = rr.statemachine.doAction() {
				errCh <- err
			}

			rr.statemachine.init(mst, &mctx)
		}
	}

}

func (rr *routeRunner) shutdown() {

}
