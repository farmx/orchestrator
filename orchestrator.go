package orchestrator

type orchestrator struct {
	defaultRoute *route
	ck           caretaker
}

func NewOrchestrator() *orchestrator {
	return &orchestrator{
		defaultRoute: newRoute("default"),
		ck:           &fileCaretaker{},
	}
}

func (o *orchestrator) addProcess(step TransactionStep) {
	o.defaultRoute.AddNextStep(step)
}

func (o *orchestrator) exec() {
	for o.defaultRoute.hasNext() {
		err := o.defaultRoute.execNextStep()
		if err != nil {
			println(err.Error())
		}

		mem := o.defaultRoute.createMemento()
		o.ck.persist(o.defaultRoute.id, mem)
	}
}
