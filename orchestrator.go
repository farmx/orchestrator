package orchestrator

//  TODO: Inject chain undoAction strategy in each atomic route handler

// Every route is created from multiple state those are connected with and edge
// Each edge has a priority and a condition
// To go to the next step the edge sorted by priority and the first one which comply the condition, the doAction continued then
// In this scenario retry and backoff algorithm can be define as a edge which it's priority will be decrease with each time execution
// Orchestrator handover context between registered route based on their identifier
type orchestrator struct {
	routes map[string]*route
}

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		routes: make(map[string]*route),
	}
}

func (o *orchestrator) register(id string, route *route) {

}

func (o *orchestrator) exec(ctxChan chan context) {
	//root := o.handlers[0]
	//errChan := make(chan error)
	//
	//for ctx := range ctxChan {
	//	ts := root.run(ctx, errChan)
	//
	//}

}

func (o *orchestrator) shutdown() error {
	return nil
}
