package orchestrator

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

// Every TransactionalRoute is created from multiple state those are connected with and edge
// Each edge has a priority and a condition
// To go to the next step the edge sorted by priority and the first do-Action which comply with the condition called
// In this scenario retry and backoff algorithm can be define as a edge which it's priority will be decrease with each time execution
// Orchestrator handover context between registered TransactionalRoute, based on their identifier
type orchestrator struct {
	// registered routes
	routes map[string]Route

	// TransactionalRoute handler
	rh *routeRunner

	ec chan error
}

type TransactionalStep interface {
	DoAction(ctx *context) error
	UndoAction(ctx context)
}

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		routes: make(map[string]Route),
	}
}

func (o *orchestrator) register(id string, r Route) error {
	if o.routes[id] != nil {
		return errors.New(fmt.Sprintf("duplicate route id %s", id))
	}

	o.routes[id] = r
	return nil
}

func (o *orchestrator) execPreparation() error {
	for _, s := range o.routes {
		for _, e := range s.GetEndpoints() {
			if o.routes[e.To] == nil {
				return errors.New(fmt.Sprintf("route id %s not found", e.To))
			}

			e.State.transitions = append(e.State.transitions, transition{
				to:       o.routes[e.To].GetStartState(),
				priority: Default,
				shouldTakeTransition: func(ctx context) bool {
					return true
				},
			})

			// transaction support
			if reflect.TypeOf(o.routes[e.To]).String() == "TransactionalRoute" {
				o.routes[e.To].GetStartState().transitions = append(o.routes[e.To].GetStartState().transitions, transition{
					to:       e.State,
					priority: Default,
					shouldTakeTransition: func(ctx context) bool {
						return ctx.GetVariable(SMStatusHeaderKey) == SMRollback
					},
				})
			}
		}
	}

	return nil
}

func (o *orchestrator) Exec(from string, ctx *context, errCh chan error) {
	if o.routes[from] == nil {
		log.Fatalf("route %s not found", from)
	}

	if err := o.execPreparation(); err != nil {
		log.Fatalf(err.Error())
	}

	o.ec = errCh
	rh := newRouteRunner(o.routes[from].GetStartState(), nil)

	rh.exec(ctx, o.ec)
}

func (o *orchestrator) shutdown() error {
	return nil
}
