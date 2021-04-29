package orchestrator

type transactionState string
type transactionStatus string

const (
	Start      transactionState = "START"
	InProgress transactionState = "IN_PROGRESS"
	Rollback   transactionState = "ROLLBACK"
	Closed     transactionState = "CLOSED"
)

const (
	Unknown transactionStatus = "UNKNOWN"
	Success transactionStatus = "SUCCESS"
	Fail    transactionStatus = "FAIL"
)

type TransactionalStep interface {
	process(ctx *context) error
	failed(ctx context)
}


type exitConditionStep struct {
	TransactionalStep
}

func (ecs *exitConditionStep) process(ctx *context) error {
	return nil
}

func (ecs *exitConditionStep) failed(ctx context) {
	// empty
}