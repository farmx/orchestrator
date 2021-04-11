package orchestrator

type orchestrator struct {
	defaultRoute *route
	ck           caretaker
}

// TODO: Inject chain failed strategy in each route
// TODO: Condition
// 			route A end --> condition --(yes)--> route B
//					                 \
// 						              ---(No)--> route C
// Restore route last State on warm-up
// Graceful shutdown
func NewOrchestrator() *orchestrator {
	ck, _ := NewFileCareTacker(".")
	return &orchestrator{
		defaultRoute: newRoute("default"),
		ck:           ck,
	}
}

func (o *orchestrator) addProcess(step TransactionStep) {
	o.defaultRoute.AddNextStep(step)
}

func (o *orchestrator) exec() error {
	if err := o.restoreLastState(); err != nil {
		ctx, _ := NewContext(nil)
		o.defaultRoute.initContext(*ctx)
	}

	for o.defaultRoute.hasNext() {
		if err := o.defaultRoute.execNextStep(); err != nil {
			return err
		}

		mem := o.defaultRoute.createMemento()
		if err := o.ck.persist(o.defaultRoute.id, mem); err != nil {
			return err
		}
	}

	return nil
}

func (o *orchestrator) restoreLastState() error {
	mem, err := o.ck.get(o.defaultRoute.id)
	if err != nil {
		return err
	}

	return o.defaultRoute.restore(mem)
}
