package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func doActionTest(ctx *context) error {
	if ctx.GetVariable("HK") == nil {
		ctx.SetVariable("HK", 0)
	}

	ctx.SetVariable("HK", ctx.GetVariable("HK").(int)+1)
	return nil
}

func undoActionTest(ctx context) error {
	return nil
}

func execTestRoute(route *State) *routeRunner {
	runner := newRouteRunner(route, nil)
	ctx, _ := NewContext()

	runner.run(ctx, nil)
	return runner
}

func TestDefineUnconditionalTransactionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		AddNextStep("2", doActionTest, undoActionTest).
		AddNextStep("3", doActionTest, undoActionTest)

	rr := execTestRoute(r.GetStartState())

	assert.Equal(t, 3, rr.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionalTransactionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		AddNextStep("2", doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", doActionTest, undoActionTest).
		AddNextStep("when_2", doActionTest, undoActionTest).
		AddNextStep("when_3", doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))

	rf := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		AddNextStep("2", doActionTest, undoActionTest).
		When(func(ctx context) bool { return false }).
		AddNextStep("when_1", doActionTest, undoActionTest).
		AddNextStep("when_2", doActionTest, undoActionTest).
		AddNextStep("when_3", doActionTest, undoActionTest)

	rh = execTestRoute(rf.GetStartState())

	assert.Equal(t, 2, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineNestedConditionalTransactionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		AddNextStep("2", doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_when_1", doActionTest, undoActionTest).
		AddNextStep("when_when_2", doActionTest, undoActionTest).
		AddNextStep("when_when_3", doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 6, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseTransactionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		AddNextStep("2", doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", doActionTest, undoActionTest).
		AddNextStep("when_2", doActionTest, undoActionTest).
		Otherwise().
		AddNextStep("otherwise_1", doActionTest, undoActionTest).
		AddNextStep("otherwise_2", doActionTest, undoActionTest).
		AddNextStep("otherwise_3", doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseAndEndTransactionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		AddNextStep("2", doActionTest, undoActionTest).
		When(func(ctx context) bool { return false }).
		AddNextStep("condition_1", doActionTest, undoActionTest).
		Otherwise().
		AddNextStep("otherwise_1", doActionTest, undoActionTest).
		AddNextStep("otherwise_2", doActionTest, undoActionTest).
		AddNextStep("otherwise_3", doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineTransactionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("condition_1", doActionTest, undoActionTest).
		AddNextStep("condition_2", doActionTest, undoActionTest).
		Otherwise().
		AddNextStep("otherwise_1", doActionTest, undoActionTest).
		End().
		AddNextStep("2", doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}
