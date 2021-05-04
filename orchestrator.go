package orchestrator

import (
	"errors"
	"fmt"
	"log"
)

// Every route is created from multiple state those are connected with and edge
// Each edge has a priority and a condition
// To go to the next step the edge sorted by priority and the first do-Action which comply with the condition called
// In this scenario retry and backoff algorithm can be define as a edge which it's priority will be decrease with each time execution
// Orchestrator handover context between registered route, based on their identifier
type orchestrator struct {
	// registered routes
	routes map[string]*route

	// latest define route
	lr *route

	// route handler
	rh *routeHandler

	// error handler
	eh *errorHandler
}

type TransactionalStep interface {
	DoAction(ctx *context) error
	UndoAction(ctx context)
}

func NewOrchestrator() *orchestrator {
	o := &orchestrator{
		routes: make(map[string]*route),
		eh:     &errorHandler{},
	}

	// TODO: remove
	go o.eh.handler()

	return o
}

func (o *orchestrator) From(from string) *orchestrator {
	if o.routes[from] != nil {
		log.Fatalf("duplicate route id %s", from)
	}

	o.routes[from] = newRoute()
	o.lr = o.routes[from]
	return o
}

func (o *orchestrator) AddStep(step TransactionalStep) *orchestrator {
	o.lr.addNextStep(step.DoAction, step.UndoAction)

	return o
}

func (o *orchestrator) When(predicate func(ctx context) bool, step TransactionalStep) *orchestrator {
	o.lr.when(predicate, step.DoAction, step.UndoAction)

	return o
}

func (o *orchestrator) Otherwise(step TransactionalStep) *orchestrator {
	o.lr.otherwise(step.DoAction, step.UndoAction)

	return o
}

func (o *orchestrator) End(step TransactionalStep) *orchestrator {
	o.lr.end(step.DoAction, step.UndoAction)

	return o
}

func (o *orchestrator) To(to string) *orchestrator {
	o.lr.addNextStep(func(ctx *context) error {
		return o.notifier(ctx, to)
	}, func(ctx context) {
		// empty
	})

	return o
}

func (o *orchestrator) notifier(ctx *context, endpoint string) error {
	if o.routes[endpoint] == nil {
		return errors.New(fmt.Sprintf("endpoint %s does not exits", endpoint))
	}

	o.rh = newRouteHandler(o.routes[endpoint].getRouteStateMachine(), nil)
	o.rh.exec(ctx, o.eh.errCh)

	return nil
}

func (o *orchestrator) exec(from string, ctx *context, errCh chan error) {
	if o.routes[from] == nil {
		log.Fatalf("route does not exists")
	}

	rh := newRouteHandler(o.routes[from].getRouteStateMachine(), nil)
	rh.exec(ctx, errCh)
}

func (o *orchestrator) shutdown() error {
	return nil
}
