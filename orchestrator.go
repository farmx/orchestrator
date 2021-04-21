package orchestrator

//  TODO: Inject chain failed strategy in each route
/* Condition
 * 			route A end --> condition --(yes)--> route B
 *					                 \
 * 						              ---(No)--> route C
 */

type condition interface {
	condition(ctx context) bool
}

type orchestrator struct {
	//route id and route pointer map
	routes map[string]*route

	// router runner pointer
	routeRunner *routeRunner

	// keep executed route id
	execPath []string
}

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		routes: make(map[string]*route),
	}
}

func (o *orchestrator) addProcess(step TransactionStep) *orchestrator {

	return o
}

func (o *orchestrator) choice() *orchestrator {
	// make graph
	return o
}

func (o *orchestrator) when(condition condition)  {

}

func (o *orchestrator) exec() {

}

func (o *orchestrator) shutdown() error {
	return o.routeRunner.shutdown()
}
