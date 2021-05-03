package orchestrator

type errorHandler struct {
	errCh chan error
}

func (eh *errorHandler) handler() {
	msg := <- eh.errCh
	println(msg.Error())
}
