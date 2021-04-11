package orchestrator

import (
	"encoding/json"
	"errors"
)

type transactionState string
type transactionStatus string

const (
	Start      transactionState = "START"
	InProgress transactionState = "IN_PROGRESS"
	Rollback   transactionState = "ROLLBACK"
	Closed     transactionState = "CLOSED"
)

const (
	Unknown transactionStatus = "UNKNOWN"
	Success transactionStatus = "SUCCESS"
	Fail    transactionStatus = "FAIL"
)

type route struct {
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

type TransactionStep interface {
	process(ctx *context) error
	failed(ctx context)
}

func newRoute(routeId string) *route {
	return &route{
		id:          routeId,
		steps:       []TransactionStep{},
		currentStep: 0,
	}
}

func (r *route) AddNextStep(step TransactionStep) *route {
	r.steps = append(r.steps, step)
	return r
}

func (r *route) initContext(ctx context) error {
	if r.state == InProgress || r.state == Rollback {
		return errors.New("can not change the context middle of the process")
	}

	r.status = Unknown
	r.state = Start
	r.ctx = &ctx

	return nil
}

func (r *route) hasNext() bool {
	if r.state == Closed {
		return false
	}

	if r.state == Start {
		return true
	}

	processNotFinished := r.state == InProgress && r.currentStep < len(r.steps)
	rollbackNotFinished := r.state == Rollback && r.currentStep >= 0

	return processNotFinished || rollbackNotFinished
}

func (r *route) execNextStep() (err error) {
	switch r.state {
	case Start, InProgress:
		if err = r.process(); err != nil {
			r.state = Rollback
			// On process failed, the previous step failed method must be called
			r.currentStep--
		}
	case Rollback:
		r.rollback()
	}

	r.updateState()
	return err
}

func (r *route) updateState() {
	if r.currentStep == len(r.steps) {
		r.status = Success
		r.state = Closed
	}

	if r.currentStep < 0 {
		r.status = Fail
		r.state = Closed
	}
}

func (r *route) process() error {
	r.state = InProgress

	if err := r.steps[r.currentStep].process(r.ctx); err != nil {
		return err
	}

	r.currentStep++
	return nil
}

func (r *route) rollback() {
	r.state = Rollback

	if r.currentStep >= 0 {
		r.steps[r.currentStep].failed(*r.ctx)
		r.currentStep--
	}
}

func (r *route) createMemento() string {
	memento := &routeSnapshot{
		RouteId: r.id,
		Step:    r.currentStep,
		State:   r.state,
		Status:  r.status,
		Context: *r.ctx,
	}

	data, _ := json.Marshal(memento)
	return string(data)
}

func (r *route) restore(memento string) error {
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
		return errors.New("invalid route State")
	}

	r.id = mem.RouteId
	r.currentStep = mem.Step
	r.state = mem.State
	r.status = mem.Status

	return nil
}
