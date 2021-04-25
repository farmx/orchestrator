package orchestrator

type atomicRouteHandler struct {
	ar *atomicRoute
	ck caretaker
}

func NewRouteHandler(atomicRoute *atomicRoute) *atomicRouteHandler {
	ck, _ := NewFileCareTacker(atomicRoute.id)
	return &atomicRouteHandler{
		ar: atomicRoute,
		ck: ck,
	}
}

// TODO: retry strategy
func (h *atomicRouteHandler) exec(errChan chan error) transactionStatus {
	for h.ar.hasNext() {
		if err := h.ar.execNextStep(); err != nil {
			errChan <- err
		}

		mem := h.ar.createMemento()
		if err := h.ck.persist(h.ar.id, mem); err != nil {
			errChan <- err
		}
	}

	return h.ar.status
}

func (h *atomicRouteHandler) run(ctx context, errChan chan error) transactionStatus {
	if err := h.ar.init(ctx); err != nil {
		errChan <- err
	}

	return h.exec(errChan)
}

func (h *atomicRouteHandler) warmUp(errChan chan error) {
	if err := h.restoreLastState(); err != nil {
		// warning log
		return
	}

	h.exec(errChan)
}

func (h *atomicRouteHandler) restoreLastState() error {
	mem, err := h.ck.get(h.ar.id)
	if err != nil {
		return err
	}

	return h.ar.restore(mem)
}

func (h *atomicRouteHandler) shutdown() error {
	_ = h.ck.persist(h.ar.id, h.ar.createMemento())
	return h.ck.shutdown()
}
