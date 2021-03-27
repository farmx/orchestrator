package orchestrator

import (
	"fmt"
	"strings"
)

const (
	transactionStateEventKey = "TRANSACTION_SATE_EVENT"

	inProgress = "IN_PROGRESS"
	rollback   = "ROLLBACK"
	done       = "DONE"
)

type route struct {
	id          string
	journald    journald
	steps       []transactionStep
	currentStep int
	status      string
}

type transactionStep interface {
	success(ctx *context) error
	failed(ctx context)
}

func newRoute(routeId string) *route {
	return &route{
		id:          routeId,
		journald:    getJournaldInstance(),
		steps:       []transactionStep{},
		currentStep: 0,
		status:      inProgress,
	}
}

func (r *route) AddNextStep(step transactionStep) *route {
	r.steps = append(r.steps, step)
	return r
}

func (r *route) Execute(ctx context) error {
	err := r.process(&ctx)

	if err != nil {
		r.rollback(ctx)
	}

	return err
}

func (r *route) process(ctx *context) error {
	for ; r.currentStep < len(r.steps); r.currentStep++ {

		if err := r.steps[r.currentStep].success(ctx); err != nil {
			return err
		}

		r.logState(*ctx)
	}

	r.closeTransaction(*ctx)
	return nil
}

func (r *route) rollback(ctx context) {
	r.status = rollback

	for ; r.currentStep >= 0; r.currentStep-- {
		r.steps[r.currentStep].failed(ctx)
		r.logState(ctx)
	}
}

func (r *route) recoverLastState(gid string) error {
	processKey := r.keyGen(transactionStateEventKey, gid)
	data, err := r.journald.getLastEvent(processKey)
	if err != nil {
		return err
	}

	r.status = fmt.Sprintf("%v", data[0])
	if r.status == done {
		return nil
	}

	r.currentStep = data[1].(int)
	ctx := data[2].(context)

	switch r.status {
	case inProgress:
		return r.process(&ctx)
	case rollback:
		r.rollback(ctx)
	}

	return nil
}

func (r *route) keyGen(primary string, secondary string) string {
	return strings.Join([]string{primary, secondary}, "_")
}

func (r *route) logState(ctx context) {
	processKey := r.keyGen(transactionStateEventKey, ctx.getGuid())
	r.journald.append(processKey, r.status, r.currentStep, ctx)
}

func (r *route) closeTransaction(ctx context) {
	processKey := r.keyGen(transactionStateEventKey, ctx.getGuid())
	r.journald.append(processKey, done, r.currentStep)
}
