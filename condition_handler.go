package orchestrator

type conditionHandler struct {
	condition condition
}

func NewConditionHandler(condition condition) *conditionHandler {
	return &conditionHandler{
		condition: condition,
	}
}

func (ch *conditionHandler) run(ctx context, errChan chan error) transactionStatus {
	if ch.condition.condition(ctx) {
		return Success
	}

	return Fail
}