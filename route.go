package orchestrator

type transactionState string

const (
	Start      transactionState = "START"
	InProgress transactionState = "IN_PROGRESS"
	Rollback   transactionState = "ROLLBACK"
	Closed     transactionState = "CLOSED"
)

type route struct {
	id          string
	journald    journald
	steps       []transactionStep
	currentStep int
	state       transactionState
}

type transactionStep interface {
	success(ctx *context) error
	failed(ctx context)
}

func newRoute(routeId string) *route {
	return &route{
		id:          routeId,
		journald:    getFileJournaldInstance(routeId),
		steps:       []transactionStep{},
		currentStep: 0,
	}
}

func (r *route) AddNextStep(step transactionStep) *route {
	r.steps = append(r.steps, step)
	return r
}

func (r *route) Execute(ctx context) error {
	r.startTransaction(ctx)

	err := r.process(&ctx)

	if err != nil {
		r.rollback(ctx)
	}

	r.closeTransaction(ctx)
	return err
}

func (r *route) process(ctx *context) error {
	for ; r.currentStep < len(r.steps); r.currentStep++ {

		if err := r.steps[r.currentStep].success(ctx); err != nil {
			return err
		}

		r.logState(*ctx)
	}

	return nil
}

func (r *route) rollback(ctx context) {
	r.state = Rollback

	for ; r.currentStep >= 0; r.currentStep-- {
		r.steps[r.currentStep].failed(ctx)
		r.logState(ctx)
	}
}

func (r *route) RecoverLastState() error {
	data, err := r.journald.getLastEvent()
	if err != nil {
		return err
	}

	r.state = data[0].(transactionState)
	if r.state == Closed {
		return nil
	}

	r.currentStep = data[1].(int)
	ctx := data[2].(context)

	switch r.state {
	case Start:
	case InProgress:
		return r.Execute(ctx)
	case Rollback:
		r.rollback(ctx)
	}

	return nil
}

func (r *route) Shutdown() {
	r.journald.shutdown()
}

func (r *route) logState(ctx context) {
	r.journald.appendLog(r.state, r.currentStep, ctx)
}

func (r *route) startTransaction(ctx context) {
	r.state = Start

	r.logState(ctx)
}

func (r *route) closeTransaction(ctx context) {
	r.state = Closed

	r.logState(ctx)
}
