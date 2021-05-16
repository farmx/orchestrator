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

func (drr *defaultRecoveryRoute) GetStartState() *state {
	return &state{
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

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		routes: make(map[string]Route),
	}
}

func (o *orchestrator) Register(r Route) error {
	if o.routes[r.GetRouteId()] != nil {
		return errors.New(fmt.Sprintf("duplicate route id %s", r.GetRouteId()))
	}

	o.routes[r.GetRouteId()] = r
	return nil
}

func (o *orchestrator) defineHierarchicalRouteTransitions() error {
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

			if reflect.TypeOf(o.routes[e.To]) == reflect.TypeOf(&TransactionalRoute{}) {
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

// "RECOVERY_ROUTE" reserved route id for recovery route
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

func (o *orchestrator) Initialization(recoveryRoute Route) error {
	if err := o.defineRecoveryRoute(recoveryRoute); err != nil {
		return err
	}

	return o.defineHierarchicalRouteTransitions()
}

// Exec connect all endpoints to the proper route according to the route id with a transition
// from is the starter route id
func (o *orchestrator) Exec(from string, ctx *context, errCh chan error) {
	if o.routes[from] == nil {
		log.Fatalf("route %s not found", from)
	}

	o.ec = errCh
	rh := newRouteRunner(o.routes[from].GetStartState(), o.routes[DefaultRecoveryRouteId].GetStartState())

	rh.exec(ctx, o.ec)
}

func (o *orchestrator) shutdown() error {
	return nil
}
