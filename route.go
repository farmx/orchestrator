package orchestrator

import "strings"

const (
	eventLogJournalKey = "EVENT_LOG"
	stepsJournalKey = "STEPS"

	success = "SUCCESS"
	failed = "FAILED"
)

type route struct {
	id string
	journald journald
	steps []transactionStep
}

type transactionStep interface{
	success(ctx *context) error
	failed(ctx context)
}

func newRoute(routeId string) *route {
	return &route{
		id: routeId,
		journald: getJournaldInstance(),
		steps: []transactionStep{},
	}
}

func (r *route) keyGen(primary string, secondary string) string {
	return strings.Join([]string{primary, secondary}, "_")
}

func (r *route) AddNextStep(step transactionStep) *route {
	r.steps = append(r.steps, step)
	return r
}

func (r *route) done()  {
	r.journald.journal(r.keyGen(r.id, stepsJournalKey), r.steps)
}

func (r *route) Execute(ctx context) (err error) {
	failedStep := -1
	
	for i, step := range r.steps {
		r.journald.journal(r.keyGen(ctx.guid, eventLogJournalKey), success, i, ctx)
		if err = step.success(&ctx); err != nil {
			failedStep = i
			break
		}
	}
	
	for i := failedStep - 1 ; i >= 0 ; i-- {
		r.steps[i].failed(ctx)
		r.journald.journal(r.keyGen(ctx.guid, eventLogJournalKey), failed, i, ctx)
	}

	return err
}

func (r *route) recoverLastState() error {
	r.journald.getLastEvent(r.keyGen("1", eventLogJournalKey))
	return nil
}