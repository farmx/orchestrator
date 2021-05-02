package orchestrator

import "testing"

type passStep struct {
	TransactionalStep
}

func (ps *passStep) doAction(ctx *context) error {
	println("doAction")
	return nil
}

func (ps *passStep) undoAction(ctx context) {
	println("undoAction")
}

func TestSample(t *testing.T) {
	/*
		orch := NewOrchestrator()
		orch.addProcess(&passStep{}).
			when().
			addProcess().
			addProcess().
			elseThen().
			addProcess().
			end().end()

		orch.exec()
	*/

	//if err := orch.exec(); err != nil {
	//	t.Fail()
	//}
}
