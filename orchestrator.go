package orchestrator

import (
	"github.com/google/uuid"
)

//  TODO: Inject chain failed strategy in each atomic route handler
/*
 * orchestrator execute a route
 * each route consist of multiple atomic route
 *
 * Condition
 * 			atomicRoute A end --> condition --(yes)--> atomicRoute B
 *					                 \
 * 						              ---(No)--> atomicRoute C
 */

type condition interface {
	condition(ctx context) bool
}

type orchestrator struct {
	// handler list
	handlers []handler

	// handler edges
	handlerEdges map[handler][]handler

	conditionHandlerStack []handler

	// latest added atomicRoute
	lastAtomicRoute *atomicRoute

	// keep executed atomicRoute id
	executedPath []string
}

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		handlerEdges: make(map[handler][]handler),
	}
}

func (o *orchestrator) addNewAtomicRoute() {
	routeId := uuid.New().String()
	r := newRoute(routeId)
	o.lastAtomicRoute = r
}

func (o *orchestrator) closeAtomicRoute(parent handler) {
	rh := NewRouteHandler(o.lastAtomicRoute)
	o.handlers = append(o.handlers, rh)

	if parent == nil || len(o.handlers) < 2 {
		return
	}

	o.defineEdge(parent, rh)
}

func (o *orchestrator) defineEdge(a, b handler) {
	o.handlerEdges[b] = append(o.handlerEdges[b], a)
	o.handlerEdges[a] = append(o.handlerEdges[a], b)
}

func (o *orchestrator) addProcess(step TransactionStep) *orchestrator {
	if o.lastAtomicRoute == nil {
		o.addNewAtomicRoute()
	}

	o.lastAtomicRoute.addNextStep(step)
	return o
}

func (o *orchestrator) when(condition condition) *orchestrator {
	prevArParent := o.handlers[len(o.handlers) - 1]
	o.closeAtomicRoute(prevArParent)

	ch := NewConditionHandler(condition)
	o.handlers = append(o.handlers, ch)
	o.conditionHandlerStack = append(o.conditionHandlerStack, ch)

	// condition pass route
	o.addNewAtomicRoute()
	return o
}

func (o *orchestrator) otherwise() *orchestrator {
	prevArParent := o.conditionHandlerStack[len(o.conditionHandlerStack) - 1]
	o.closeAtomicRoute(prevArParent)
	o.addNewAtomicRoute()
	return o
}

func (o *orchestrator) whenEnd() *orchestrator {
	prevArParent := o.conditionHandlerStack[len(o.conditionHandlerStack) - 1]
	// pop conditionHandlerStack
	o.conditionHandlerStack = o.conditionHandlerStack[:len(o.conditionHandlerStack) - 1]

	o.closeAtomicRoute(prevArParent)
	o.addNewAtomicRoute()

	return o
}

func (o *orchestrator) exec(ctxChan chan context) {

}

func (o *orchestrator) shutdown() error {
	return nil
}
