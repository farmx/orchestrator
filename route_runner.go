package orchestrator

type routeRunner struct {
	r            *route
	ck           caretaker
	errorHandler errorHandler
}

func NewRouteRunner(route *route) *routeRunner {
	ck, _ := NewFileCareTacker(route.id)
	return &routeRunner{
		r:  route,
		ck: ck,
	}
}

// TODO: retry strategy
func (rr *routeRunner) exec(errChan chan error) transactionStatus {
	for rr.r.hasNext() {
		if err := rr.r.execNextStep(); err != nil {
			errChan <- err
		}

		mem := rr.r.createMemento()
		if err := rr.ck.persist(rr.r.id, mem); err != nil {
			errChan <- err
		}
	}

	return rr.r.status
}

// Restore route last State on warm-up
func (rr *routeRunner) run(ctxChan chan context, errChan chan error) {
	if err := rr.restoreLastState(); err != nil {
		// warning log
	} else {
		rr.exec(errChan)
	}

	for ctx := range ctxChan {
		if err := rr.r.init(ctx); err != nil {
			errChan <- err
			continue
		}

		rr.exec(errChan)
	}
}

func (rr *routeRunner) restoreLastState() error {
	mem, err := rr.ck.get(rr.r.id)
	if err != nil {
		return err
	}

	return rr.r.restore(mem)
}

func (rr *routeRunner) shutdown() error {
	_ = rr.ck.persist(rr.r.id, rr.r.createMemento())
	return rr.ck.shutdown()
}
