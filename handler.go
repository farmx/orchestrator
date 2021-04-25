package orchestrator

type handler interface {
	run(ctx context, errChan chan error) transactionStatus
}
