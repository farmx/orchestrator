package orchestrator

import "testing"

type passStep struct {
	TransactionalStep
}

func (ps *passStep) process(ctx *context) error {
	println("process")
	return nil
}

func (ps *passStep) failed(ctx context) {
	println("failed")
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
