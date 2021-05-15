package orchestrator

import (
	"errors"
	"fmt"
	"log"
)

// Every transactionalRoute is created from multiple state those are connected with and edge
// Each edge has a priority and a condition
// To go to the next step the edge sorted by priority and the first do-Action which comply with the condition called
// In this scenario retry and backoff algorithm can be define as a edge which it's priority will be decrease with each time execution
// Orchestrator handover context between registered transactionalRoute, based on their identifier
type orchestrator struct {
	// registered routes
	routes map[string]*transactionalRoute

	// latest define transactionalRoute
	lr *transactionalRoute

	// transactionalRoute handler
	rh *routeHandler

	ec chan error
}

type TransactionalStep interface {
	DoAction(ctx *context) error
	UndoAction(ctx context)
}

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		routes: make(map[string]*transactionalRoute),
	}
}

func (o *orchestrator) From(from string) *orchestrator {
	if o.routes[from] != nil {
		log.Fatalf("duplicate transactionalRoute id %s", from)
	}

	o.routes[from] = newRoute()
	o.lr = o.routes[from]
	return o
}

func (o *orchestrator) AddStep(step TransactionalStep) *orchestrator {
	o.lr.addNextStep(step.DoAction, step.UndoAction)

	return o
}

func (o *orchestrator) When(predicate func(ctx context) bool) *orchestrator {
	o.lr.when(predicate)

	return o
}

func (o *orchestrator) Otherwise() *orchestrator {
	o.lr.otherwise()

	return o
}

func (o *orchestrator) End() *orchestrator {
	o.lr.end()

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
	o.rh.exec(ctx, o.ec)

	return nil
}

func (o *orchestrator) Exec(from string, ctx *context, errCh chan error) {
	if o.routes[from] == nil {
		log.Fatalf("transactionalRoute does not exists")
	}

	o.ec = errCh
	rh := newRouteHandler(o.routes[from].getRouteStateMachine(), nil)

	rh.exec(ctx, o.ec)
}

func (o *orchestrator) shutdown() error {
	return nil
}
