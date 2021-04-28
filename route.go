package orchestrator

import (
	"encoding/json"
	"errors"
)

type priority int

const (
	Condition priority = 2
	Default   priority = 1
)

type (
	route struct {
		id             string
		steps          []TransactionalStep
		transitions    map[TransactionalStep][]transition
		conditionStack []TransactionalStep
		currentStep    int
		state          transactionState
		ctx            *context
		status         transactionStatus
	}

	transition struct {
		to                   TransactionalStep
		priority             priority
		shouldTakeTransition func(ctx context) bool
	}

	routeSnapshot struct {
		RouteId string
		Context context
		Step    int
		Status  transactionStatus
		State   transactionState
	}
)

func NewRoute(routeId string) *route {
	return &route{
		id:          routeId,
		steps:       []TransactionalStep{},
		transitions: make(map[TransactionalStep][]transition),
		currentStep: 0,
	}
}

func (r *route) AddNextStep(step TransactionalStep) {
	r.steps = append(r.steps, step)

	if len(r.steps) > 1 {
		ps := r.steps[len(r.steps)-2]
		cs := r.steps[len(r.steps)-1]

		r.transitions[ps] = append(r.transitions[ps], transition{
			to:       cs,
			priority: Default,
			shouldTakeTransition: func(ctx context) bool {
				return true
			},
		})
	}
}

func (r *route) When(condition func(ctx context) bool, step TransactionalStep) {
	r.steps = append(r.steps, step)

	ps := r.steps[len(r.steps)-2]
	cs := r.steps[len(r.steps)-1]

	// Store conditional transition source for otherwise part
	r.conditionStack = append(r.conditionStack, ps)

	r.transitions[ps] = append(r.transitions[ps], transition{
		to:                   cs,
		priority:             Condition,
		shouldTakeTransition: condition,
	})
}

func (r *route) Otherwise(step TransactionalStep) {
	r.steps = append(r.steps, step)

	// Pop last condition transition source
	s := r.conditionStack[len(r.conditionStack)-1]
	r.conditionStack = r.conditionStack[:len(r.conditionStack)-1]

	// Find the condition transition
	for _, t := range r.transitions[s] {
		if t.priority == Condition {
			r.transitions[s] = append(r.transitions[s], transition{
				to:       step,
				priority: Condition,
				shouldTakeTransition: func(ctx context) bool {
					return !t.shouldTakeTransition(ctx)
				},
			})

			return
		}
	}
}

func (r *route) init(ctx context) error {
	if r.state == InProgress || r.state == Rollback {
		return errors.New("route in progress. can not change the context")
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
