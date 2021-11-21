package orchestrator

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

// Every TransactionalRoute is created from multiple State those are connected with and edge
// Each edge has a priority and a condition
// To go to the doAction step the edge sorted by priority and the first do-Action which comply with the condition called
// In this scenario retry and backoff algorithm can be define as a edge which it's priority will be decrease with each time execution
// Orchestrator handover context between registered TransactionalRoute, based on their identifier

const DefaultRecoveryRouteId = "RECOVERY_ROUTE"

type (
	orchestrator struct {
		// registered routes
		routes map[string]Route

		// TransactionalRoute handler
		rh *routeRunner

		ec chan error
	}

	defaultRecoveryRoute struct {
	}
)

func (drr *defaultRecoveryRoute) GetRouteId() string {
	return "RECOVERY_ROUTE"
}

func (drr *defaultRecoveryRoute) GetStartState() *State {
	return &State{
		name:        "default_recovery_state",
		transitions: nil,
		action: func(ctx *context) error {
			return nil
		},
	}
}

func (drr *defaultRecoveryRoute) GetEndpoints() []*Endpoint {
	return nil
}

// NewOrchestrator create and init orchestrator
func NewOrchestrator() *orchestrator {
	return &orchestrator{
		routes: make(map[string]Route),
	}
}

// Register is for register a route with it's unique identifier
func (o *orchestrator) Register(r Route) error {
	if o.routes[r.GetRouteId()] != nil {
		return errors.New(fmt.Sprintf("duplicate route id %s", r.GetRouteId()))
	}

	if r.GetRouteId() == DefaultRecoveryRouteId {
		return errors.New(DefaultRecoveryRouteId + " route id is reserved")
	}

	o.routes[r.GetRouteId()] = r
	return nil
}

// Initialization define recovery route and define transition between routes (HierarchicalRoute feature)
func (o *orchestrator) Initialization(recoveryRoute Route) error {
	if err := o.defineRecoveryRoute(recoveryRoute); err != nil {
		return err
	}

	return o.defineHierarchicalRouteTransitions()
}

// Exec start the execution process from the route id with a context
func (o *orchestrator) Exec(from string, ctx *context, errCh chan error) {
	if o.routes[from] == nil {
		log.Fatalf("route %s not found", from)
	}

	o.ec = errCh
	rh := newRouteRunner(o.routes[from].GetStartState(), o.routes[DefaultRecoveryRouteId].GetStartState())

	rh.run(ctx, o.ec)
}

func (o *orchestrator) defineHierarchicalRouteTransitions() error {
	for _, s := range o.routes {
		for _, e := range s.GetEndpoints() {
			if o.routes[e.To] == nil {
				return errors.New(fmt.Sprintf("route id %s not found", e.To))
			}

			e.State.createTransition(o.routes[e.To].GetStartState(), Default,
				func(ctx context) bool {
					return true
				})

			if reflect.TypeOf(o.routes[e.To]) == reflect.TypeOf(&TransactionalRoute{}) {
				o.routes[e.To].GetStartState().createTransition(e.State, Default,
					func(ctx context) bool {
						return ctx.GetVariable(transactionStatusHeaderKey) == transactionStatusRollback
					})
			}
		}
	}

	return nil
}

// "RECOVERY_ROUTE" route id is reserved for recovery route
func (o *orchestrator) defineRecoveryRoute(route Route) error {
	r := route

	if route == nil {
		r = &defaultRecoveryRoute{}
	}

	if r.GetRouteId() != DefaultRecoveryRouteId {
		return errors.New("recovery route id is not 'RECOVERY_ROUTE'")
	}

	o.routes[DefaultRecoveryRouteId] = r
	return nil
}

func (o *orchestrator) shutdown() error {
	return nil
}
