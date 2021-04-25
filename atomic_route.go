package orchestrator

import (
	"encoding/json"
	"errors"
)

type atomicRoute struct {
	id          string
	steps       []TransactionStep
	currentStep int
	state       transactionState
	ctx         *context
	status      transactionStatus
}

type routeSnapshot struct {
	RouteId string
	Context context
	Step    int
	Status  transactionStatus
	State   transactionState
}

func newRoute(routeId string) *atomicRoute {
	return &atomicRoute{
		id:          routeId,
		steps:       []TransactionStep{},
		currentStep: 0,
	}
}

func (ar *atomicRoute) addNextStep(step TransactionStep) {
	ar.steps = append(ar.steps, step)
}

func (ar *atomicRoute) init(ctx context) error {
	if ar.state == InProgress || ar.state == Rollback {
		return errors.New("atomicRoute in progress. can not change the context")
	}

	ar.status = Unknown
	ar.state = Start
	ar.ctx = &ctx

	return nil
}

func (ar *atomicRoute) hasNext() bool {
	if ar.state == Closed {
		return false
	}

	if ar.state == Start {
		return true
	}

	processNotFinished := ar.state == InProgress && ar.currentStep < len(ar.steps)
	rollbackNotFinished := ar.state == Rollback && ar.currentStep >= 0

	return processNotFinished || rollbackNotFinished
}

func (ar *atomicRoute) execNextStep() (err error) {
	switch ar.state {
	case Start, InProgress:
		if err = ar.process(); err != nil {
			ar.state = Rollback
			// On process failed, the previous step failed method must be called
			ar.currentStep--
		}
	case Rollback:
		ar.rollback()
	}

	ar.updateState()
	return err
}

func (ar *atomicRoute) updateState() {
	if ar.currentStep == len(ar.steps) {
		ar.status = Success
		ar.state = Closed
	}

	if ar.currentStep < 0 {
		ar.status = Fail
		ar.state = Closed
	}
}

func (ar *atomicRoute) process() error {
	ar.state = InProgress

	if err := ar.steps[ar.currentStep].process(ar.ctx); err != nil {
		return err
	}

	ar.currentStep++
	return nil
}

func (ar *atomicRoute) rollback() {
	ar.state = Rollback

	if ar.currentStep >= 0 {
		ar.steps[ar.currentStep].failed(*ar.ctx)
		ar.currentStep--
	}
}

func (ar *atomicRoute) createMemento() string {
	memento := &routeSnapshot{
		RouteId: ar.id,
		Step:    ar.currentStep,
		State:   ar.state,
		Status:  ar.status,
		Context: *ar.ctx,
	}

	data, _ := json.Marshal(memento)
	return string(data)
}

func (ar *atomicRoute) restore(memento string) error {
	var mem routeSnapshot
	if err := json.Unmarshal([]byte(memento), &mem); err != nil {
		return err
	}

	if mem.RouteId == "" {
		return errors.New("RouteId is empty")
	}

	if mem.Step < 0 {
		return errors.New("negative value does not allow for Step")
	}

	if mem.State != Start && mem.State != Closed && mem.State != InProgress && mem.State != Rollback {
		return errors.New("invalid atomicRoute State")
	}

	ar.id = mem.RouteId
	ar.currentStep = mem.Step
	ar.state = mem.State
	ar.status = mem.Status

	return nil
}
